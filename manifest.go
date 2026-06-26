package mcgo

import "fmt"

const manifestURL = "https://launchermeta.mojang.com/mc/game/version_manifest_v2.json"

func fetchManifest() (*VersionManifest, error) {
	var m VersionManifest
	if err := fetchJSON(manifestURL, &m); err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}
	return &m, nil
}

// Versions returns all available Minecraft versions from the Mojang API.
func Versions() ([]ManifestVersion, error) {
	m, err := fetchManifest()
	if err != nil {
		return nil, err
	}
	return m.Versions, nil
}

// Latest returns the ID of the latest release or snapshot.
func Latest(kind string) (string, error) {
	m, err := fetchManifest()
	if err != nil {
		return "", err
	}
	switch kind {
	case "release":
		return m.Latest.Release, nil
	case "snapshot":
		return m.Latest.Snapshot, nil
	}
	return "", fmt.Errorf("unknown kind: %s (use 'release' or 'snapshot')", kind)
}

// versionURL looks up the JSON URL for a specific version ID.
func versionURL(id string) (string, error) {
	versions, err := Versions()
	if err != nil {
		return "", err
	}
	for _, v := range versions {
		if v.ID == id {
			return v.URL, nil
		}
	}
	return "", fmt.Errorf("version %q not found", id)
}
