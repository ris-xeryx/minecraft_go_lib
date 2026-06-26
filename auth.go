package mcgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Authenticator is implemented by all auth methods.
type Authenticator interface {
	Login() (Profile, error)
}

// ── Offline ──────────────────────────────────────────────────────────────

type OfflineAuth struct{ Username string }

func (a OfflineAuth) Login() (Profile, error) {
	hash := md5Hash([]byte("OfflinePlayer:" + a.Username))
	return Profile{Username: a.Username, UUID: uuidFormat(hash)}, nil
}

// ── Microsoft device-code flow ───────────────────────────────────────────

type MicrosoftAuth struct {
	ClientID string
	OnCode   func(userCode, verifyURL string)
}

func (m MicrosoftAuth) Login() (Profile, error) {
	if m.ClientID == "" {
		return Profile{}, fmt.Errorf("client_id required")
	}
	dc, err := m.deviceCode()
	if err != nil {
		return Profile{}, err
	}
	if m.OnCode != nil {
		m.OnCode(dc.UserCode, dc.Verification)
	}
	ms, err := m.poll(dc)
	if err != nil {
		return Profile{}, err
	}
	xb, err := m.xboxLive(ms.AccessToken)
	if err != nil {
		return Profile{}, err
	}
	xsts, uhs, err := m.xsts(xb.Token)
	if err != nil {
		return Profile{}, err
	}
	mcTok, err := m.minecraft(xsts, uhs)
	if err != nil {
		return Profile{}, err
	}
	p, err := m.profile(mcTok)
	if err != nil {
		return Profile{}, err
	}
	return Profile{Username: p.Name, UUID: p.ID, AccessToken: mcTok}, nil
}

type dcResp struct {
	UserCode     string `json:"user_code"`
	DeviceCode   string `json:"device_code"`
	Verification string `json:"verification_uri"`
	ExpiresIn    int    `json:"expires_in"`
	Interval     int    `json:"interval"`
}

func (m MicrosoftAuth) deviceCode() (*dcResp, error) {
	body := fmt.Sprintf("client_id=%s&scope=XboxLive.signin+offline_access", m.ClientID)
	req, _ := http.NewRequest("POST",
		"https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("device_code HTTP %d", resp.StatusCode)
	}
	var dc dcResp
	json.NewDecoder(resp.Body).Decode(&dc)
	return &dc, nil
}

type tokenResp struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error,omitempty"`
	ErrorDesc   string `json:"error_description,omitempty"`
}

func (m MicrosoftAuth) poll(dc *dcResp) (*tokenResp, error) {
	sec := dc.Interval
	if sec == 0 {
		sec = 5
	}
	deadline := time.Now().Add(time.Duration(dc.ExpiresIn) * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(time.Duration(sec) * time.Second)
		body := fmt.Sprintf(
			"grant_type=urn:ietf:params:oauth:grant-type:device_code&client_id=%s&device_code=%s",
			m.ClientID, dc.DeviceCode)
		req, _ := http.NewRequest("POST",
			"https://login.microsoftonline.com/consumers/oauth2/v2.0/token",
			strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := httpClient.Do(req)
		if err != nil {
			continue
		}
		var tr tokenResp
		json.NewDecoder(resp.Body).Decode(&tr)
		resp.Body.Close()
		if tr.Error == "authorization_pending" {
			continue
		}
		if tr.Error != "" {
			return nil, fmt.Errorf("%s: %s", tr.Error, tr.ErrorDesc)
		}
		if tr.AccessToken != "" {
			return &tr, nil
		}
	}
	return nil, fmt.Errorf("device code expired")
}

type xboxResp struct {
	Token         string `json:"Token"`
	DisplayClaims struct {
		Xui []struct {
			UHS string `json:"uhs"`
		} `json:"xui"`
	} `json:"DisplayClaims"`
}

func (m MicrosoftAuth) xboxLive(accessToken string) (*xboxResp, error) {
	return xboxCall("https://user.auth.xboxlive.com/user/authenticate", map[string]any{
		"Properties": map[string]any{
			"AuthMethod": "RPS", "SiteName": "user.auth.xboxlive.com",
			"RpsTicket": "d=" + accessToken,
		},
		"RelyingParty": "http://auth.xboxlive.com", "TokenType": "JWT",
	})
}

func (m MicrosoftAuth) xsts(xboxToken string) (token, uhs string, err error) {
	xb, err := xboxCall("https://xsts.auth.xboxlive.com/xsts/authorize", map[string]any{
		"Properties": map[string]any{
			"SandboxId": "RETAIL", "UserTokens": []string{xboxToken},
		},
		"RelyingParty": "rp://api.minecraftservices.com/", "TokenType": "JWT",
	})
	if err != nil {
		return "", "", err
	}
	if len(xb.DisplayClaims.Xui) == 0 {
		return "", "", fmt.Errorf("no uhs")
	}
	return xb.Token, xb.DisplayClaims.Xui[0].UHS, nil
}

func (m MicrosoftAuth) minecraft(xsts, uhs string) (string, error) {
	body := fmt.Sprintf(`{"identityToken":"XBL3.0 x=%s;%s"}`, uhs, xsts)
	resp, err := httpClient.Post("https://api.minecraftservices.com/authentication/login_with_xbox",
		"application/json", strings.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("mc login HTTP %d", resp.StatusCode)
	}
	var tr struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(resp.Body).Decode(&tr)
	return tr.AccessToken, nil
}

func xboxCall(url string, body map[string]any) (*xboxResp, error) {
	b, _ := json.Marshal(body)
	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("xbox HTTP %d: %s", resp.StatusCode, url)
	}
	var xb xboxResp
	json.NewDecoder(resp.Body).Decode(&xb)
	return &xb, nil
}

type mcProfile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (m MicrosoftAuth) profile(accessToken string) (*mcProfile, error) {
	req, _ := http.NewRequest("GET", "https://api.minecraftservices.com/minecraft/profile", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("profile HTTP %d", resp.StatusCode)
	}
	var p mcProfile
	json.NewDecoder(resp.Body).Decode(&p)
	return &p, nil
}
