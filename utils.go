package mcgo

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// currentOSName devuelve el nombre de OS que usa Mojang en sus metadatos.
func currentOSName() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "osx"
	default:
		return "linux"
	}
}

// currentArchName devuelve la arquitectura en formato Mojang.
func currentArchName() string {
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

// httpGet con timeout default.
func httpGet(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	return resp, nil
}

// httpGetJSON hace GET y parsea JSON.
func httpGetJSON(url string, v interface{}) error {
	resp, err := httpGet(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return jsonDecode(resp.Body, v)
}

// ensureDir crea el directorio si no existe.
func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// sha1OfFile calcula el SHA1 de un archivo.
func sha1OfFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// fileExists verifica si un archivo existe.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// getenvLee una variable de entorno (wrapper para testability).
func getenv(key string) string {
	return os.Getenv(key)
}

// downloadFileDescarga un archivo (con verificación SHA1 opcional y progreso).
// name para reportar.
func downloadFile(url, path, expectedSHA1 string, bus *EventBus) error {
	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}

	if bus != nil {
		bus.emitDownloadStarted(url)
	}

	resp, err := httpGet(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	total := resp.ContentLength
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	buf := make([]byte, 32*1024)
	var loaded int64
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}
			loaded += int64(n)
			if bus != nil {
				bus.emitProgress(loaded, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	if bus != nil {
		bus.emitFileDownloaded(path)
	}

	if expectedSHA1 != "" {
		actual, err := sha1OfFile(path)
		if err != nil {
			return err
		}
		if !strings.EqualFold(actual, expectedSHA1) {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSHA1, actual)
		}
	}

	return nil
}

// downloadIfMissingDescarga si el archivo no existe o checksum inválido.
func downloadIfMissing(url, path, expectedSHA1 string, bus *EventBus) error {
	if bus != nil {
		bus.emitFileChecked(path)
	}
	if fileExists(path) {
		if expectedSHA1 == "" {
			return nil
		}
		actual, err := sha1OfFile(path)
		if err == nil && strings.EqualFold(actual, expectedSHA1) {
			return nil
		}
	}
	return downloadFile(url, path, expectedSHA1, bus)
}
