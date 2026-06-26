package mcgo

import "fmt"

const (
	manifestURL = "https://launchermeta.mojang.com/mc/game/version_manifest_v2.json"
)

// GetVersionManifest obtiene el manifiesto de versiones de Mojang.
func GetVersionManifest() (*VersionManifest, error) {
	var m VersionManifest
	if err := httpGetJSON(manifestURL, &m); err != nil {
		return nil, fmt.Errorf("fetching manifest: %w", err)
	}
	return &m, nil
}

// GetAllVersions obtiene la lista de versiones disponibles.
func GetAllVersions() ([]Version, error) {
	m, err := GetVersionManifest()
	if err != nil {
		return nil, err
	}
	return m.Versions, nil
}

// GetVersionURL obtiene la URL del JSON de una versión específica.
func GetVersionURL(versionID string) (string, error) {
	versions, err := GetAllVersions()
	if err != nil {
		return "", err
	}
	for _, v := range versions {
		if v.ID == versionID {
			return v.URL, nil
		}
	}
	return "", fmt.Errorf("version %s not found", versionID)
}

// GetLatestRelease devuelve la versión release más reciente.
func GetLatestRelease() (string, error) {
	m, err := GetVersionManifest()
	if err != nil {
		return "", err
	}
	return m.Latest.Release, nil
}

// GetLatestSnapshot devuelve la snapshot más reciente.
func GetLatestSnapshot() (string, error) {
	m, err := GetVersionManifest()
	if err != nil {
		return "", err
	}
	return m.Latest.Snapshot, nil
}
