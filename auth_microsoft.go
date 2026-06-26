package mcgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MicrosoftAuth maneja el flujo device code de Microsoft.
type MicrosoftAuth struct {
	ClientID           string
	DeviceCodeCallback func(code, url string)
}

func NewMicrosoftAuth(clientID string) *MicrosoftAuth {
	return &MicrosoftAuth{ClientID: clientID}
}

const (
	azureDeviceCodeURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode"
	azureTokenURL      = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"
	azureScope         = "https://helper.minecraftauth.net/auth.request xbox.profile.read"
)

type deviceCodeResponse struct {
	UserCode     string `json:"user_code"`
	DeviceCode   string `json:"device_code"`
	Verification string `json:"verification_uri"`
	ExpiresIn    int    `json:"expires_in"`
	Interval     int    `json:"interval"`
	Message      string `json:"message"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

// Authenticate realiza el flujo device code completo:
// 1. pide device_code al usuario
// 2. poll del token
// 3. intercambio con Xbox Live
// 4. intercambio con XSTS
// 5. intercambio con Minecraft
// 6. obtiene el profile del jugador
func (m *MicrosoftAuth) Authenticate() (Profile, error) {
	if m.ClientID == "" {
		return Profile{}, fmt.Errorf("client_id is required")
	}

	// 1. Solicitar device code
	deviceCode, err := m.requestDeviceCode()
	if err != nil {
		return Profile{}, err
	}

	// Notificar UI
	if m.DeviceCodeCallback != nil {
		m.DeviceCodeCallback(deviceCode.UserCode, deviceCode.Verification)
	}

	// 2. Poll del token
	msToken, err := m.pollForToken(deviceCode)
	if err != nil {
		return Profile{}, err
	}

	// 3. Xbox Live token
	xboxToken, err := m.getXboxToken(msToken.AccessToken, msToken.RefreshToken)
	if err != nil {
		return Profile{}, err
	}

	// 4. XSTS token
	xstsToken, uhs, err := m.getXSTSToken(xboxToken)
	if err != nil {
		return Profile{}, err
	}

	// 5. Minecraft token
	mcToken, err := m.getMinecraftToken(xstsToken, uhs)
	if err != nil {
		return Profile{}, err
	}

	// 6. Profile del jugador
	mcProfile, err := m.getMinecraftProfile(mcToken)
	if err != nil {
		return Profile{}, err
	}

	return Profile{
		Username:    mcProfile.Name,
		UUID:        mcProfile.ID,
		AccessToken: mcToken,
		PlayerName:  mcProfile.Name,
		XUID:        mcProfile.ID,
	}, nil
}

func (m *MicrosoftAuth) requestDeviceCode() (*deviceCodeResponse, error) {
	body := fmt.Sprintf("client_id=%s&scope=%s", m.ClientID, "XboxLive.signin+offline_access")
	req, err := http.NewRequest("POST", azureDeviceCodeURL, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("devicecode failed: %d", resp.StatusCode)
	}

	var dc deviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&dc); err != nil {
		return nil, err
	}
	return &dc, nil
}

func (m *MicrosoftAuth) pollForToken(dc *deviceCodeResponse) (*tokenResponse, error) {
	interval := dc.Interval
	if interval == 0 {
		interval = 5
	}
	deadline := time.Now().Add(time.Duration(dc.ExpiresIn) * time.Second)

	for time.Now().Before(deadline) {
		time.Sleep(time.Duration(interval) * time.Second)

		body := fmt.Sprintf("grant_type=urn:ietf:params:oauth:grant-type:device_code&client_id=%s&device_code=%s",
			m.ClientID, dc.DeviceCode)
		req, _ := http.NewRequest("POST", azureTokenURL, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var tr tokenResponse
		json.NewDecoder(resp.Body).Decode(&tr)

		if tr.Error == "authorization_pending" {
			continue
		}
		if tr.Error != "" {
			return nil, fmt.Errorf("auth: %s - %s", tr.Error, tr.ErrorDesc)
		}
		if tr.AccessToken != "" {
			return &tr, nil
		}
	}
	return nil, fmt.Errorf("device code expired")
}

// ── Xbox Live / XSTS / Minecraft tokens ────────────────────────────────

type xboxTokenResponse struct {
	Token         string `json:"Token"`
	DisplayClaims struct {
		Xui []struct {
			UHS string `json:"uhs"`
		} `json:"xui"`
	} `json:"DisplayClaims"`
}

func (m *MicrosoftAuth) getXboxToken(accessToken, refreshToken string) (*xboxTokenResponse, error) {
	body := map[string]interface{}{
		"Properties": map[string]interface{}{
			"AuthMethod": "RPS",
			"SiteName":   "user.auth.xboxlive.com",
			"RpsTicket":  "d=" + accessToken,
		},
		"RelyingParty": "http://auth.xboxlive.com",
		"TokenType":    "JWT",
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post("https://user.auth.xboxlive.com/user/authenticate",
		"application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("xbox auth %d: %s", resp.StatusCode, b)
	}

	var xt xboxTokenResponse
	json.NewDecoder(resp.Body).Decode(&xt)
	return &xt, nil
}

func (m *MicrosoftAuth) getXSTSToken(xboxToken *xboxTokenResponse) (string, string, error) {
	if len(xboxToken.DisplayClaims.Xui) == 0 {
		return "", "", fmt.Errorf("no uhs in xbox token")
	}
	uhs := xboxToken.DisplayClaims.Xui[0].UHS

	body := map[string]interface{}{
		"Properties": map[string]interface{}{
			"SandboxId":  "RETAIL",
			"UserTokens": []string{xboxToken.Token},
		},
		"RelyingParty": "rp://api.minecraftservices.com/",
		"TokenType":    "JWT",
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post("https://xsts.auth.xboxlive.com/xsts/authorize",
		"application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("xsts %d: %s", resp.StatusCode, b)
	}

	var xt xboxTokenResponse
	json.NewDecoder(resp.Body).Decode(&xt)
	return xt.Token, uhs, nil
}

func (m *MicrosoftAuth) getMinecraftToken(xstsToken, uhs string) (string, error) {
	body := fmt.Sprintf(`{"identityToken":"XBL3.0 x=%s;%s"}`, uhs, xstsToken)
	resp, err := http.Post("https://api.minecraftservices.com/authentication/login_with_xbox",
		"application/json", bytes.NewBufferString(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("mc login %d", resp.StatusCode)
	}

	var tr struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(resp.Body).Decode(&tr)
	return tr.AccessToken, nil
}

type minecraftProfileResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (m *MicrosoftAuth) getMinecraftProfile(accessToken string) (*minecraftProfileResponse, error) {
	req, _ := http.NewRequest("GET", "https://api.minecraftservices.com/minecraft/profile", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("profile %d", resp.StatusCode)
	}

	var p minecraftProfileResponse
	json.NewDecoder(resp.Body).Decode(&p)
	return &p, nil
}
