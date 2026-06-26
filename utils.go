package mcgo

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ── Platform detection ───────────────────────────────────────────────────

func currentOS() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "osx"
	default:
		return "linux"
	}
}

func currentArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "386":
		return "x86"
	case "arm64":
		return "arm64"
	default:
		return "x64"
	}
}

// ── HTTP ─────────────────────────────────────────────────────────────────

var httpClient = &http.Client{Timeout: 30 * time.Second}

func fetch(url string) (*http.Response, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	return resp, nil
}

func fetchJSON(url string, v any) error {
	resp, err := fetch(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

func readJSON(path string, v any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func postJSON(url string, body any, v any) error {
	b, _ := json.Marshal(body)
	resp, err := httpClient.Post(url, "application/json", strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}
	return nil
}

// ── Filesystem ───────────────────────────────────────────────────────────

func mkdir(path string) error { return os.MkdirAll(path, 0755) }

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ── SHA1 ─────────────────────────────────────────────────────────────────

func sha1File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha1.New()
	io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ── MD5 + UUID helpers ──────────────────────────────────────────────────

func md5Hash(data []byte) []byte {
	h := md5.Sum(data)
	return h[:]
}

func uuidFormat(b []byte) string {
	b[6] = (b[6] & 0x0f) | 0x30
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// ── Download ─────────────────────────────────────────────────────────────

func downloadFile(url, path, sha1 string, bus *EventBus) error {
	if err := mkdir(filepath.Dir(path)); err != nil {
		return err
	}
	if bus != nil {
		bus.emitDownloadStarted(url)
	}

	resp, err := fetch(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	total := resp.ContentLength
	buf := make([]byte, 32*1024)
	var loaded int64
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			loaded += int64(n)
			if bus != nil {
				bus.emitProgress(loaded, total)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	if bus != nil {
		bus.emitFileDownloaded(path)
	}

	if sha1 != "" {
		if actual, _ := sha1File(path); !strings.EqualFold(actual, sha1) {
			return fmt.Errorf("sha1 mismatch: %s != %s", sha1, actual)
		}
	}
	return nil
}

func downloadIfMissing(url, path, sha1 string, bus *EventBus) error {
	if bus != nil {
		bus.emitFileChecked(path)
	}
	if exists(path) {
		if sha1 == "" {
			return nil
		}
		if actual, _ := sha1File(path); strings.EqualFold(actual, sha1) {
			return nil
		}
	}
	return downloadFile(url, path, sha1, bus)
}

// ── Misc ─────────────────────────────────────────────────────────────────

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
