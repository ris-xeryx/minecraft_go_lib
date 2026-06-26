package mcgo

import (
	"fmt"
	"sort"
)

// LoaderVersion resolves the latest version of a loader for a given MC version.
func LoaderVersion(loader Loader, mcVersion string) (string, error) {
	switch loader {
	case Vanilla:
		return "", nil
	case Fabric:
		return resolveFabric(mcVersion)
	case Quilt:
		return resolveQuilt(mcVersion)
	case Forge:
		return resolveForge(mcVersion)
	case NeoForge:
		return resolveNeoForge(mcVersion)
	}
	return "", fmt.Errorf("unknown loader: %s", loader)
}

// ── Fabric ───────────────────────────────────────────────────────────────

func resolveFabric(mcVer string) (string, error) {
	var v []struct {
		Loader struct{ Version string } `json:"loader"`
	}
	if err := fetchJSON("https://meta.fabricmc.net/v2/versions/loader/"+mcVer, &v); err != nil {
		return "", err
	}
	if len(v) == 0 {
		return "", fmt.Errorf("no fabric loader for %s", mcVer)
	}
	return v[0].Loader.Version, nil
}

// ── Quilt ────────────────────────────────────────────────────────────────

func resolveQuilt(mcVer string) (string, error) {
	var v []struct {
		Loader struct{ Version string } `json:"loader"`
	}
	if err := fetchJSON("https://meta.quiltmc.org/v3/versions/loader/"+mcVer, &v); err != nil {
		return "", err
	}
	if len(v) == 0 {
		return "", fmt.Errorf("no quilt loader for %s", mcVer)
	}
	return v[0].Loader.Version, nil
}

// ── Forge ────────────────────────────────────────────────────────────────

func resolveForge(mcVer string) (string, error) {
	var v struct {
		Promos map[string]string `json:"promos"`
	}
	if err := fetchJSON("https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json", &v); err != nil {
		return "", err
	}
	if val := v.Promos[mcVer+"-recommended"]; val != "" {
		return val, nil
	}
	if val := v.Promos[mcVer+"-latest"]; val != "" {
		return val, nil
	}
	return "", fmt.Errorf("no forge for %s", mcVer)
}

// ── NeoForge ─────────────────────────────────────────────────────────────

func resolveNeoForge(mcVer string) (string, error) {
	var v struct {
		Versions []string `json:"versions"`
	}
	if err := fetchJSON("https://maven.neoforged.net/api/maven/versions/releases/net/neoforged/neoforge", &v); err != nil {
		return "", err
	}
	var parts []int
	var n int
	for _, c := range mcVer {
		if c == '.' {
			parts = append(parts, n)
			n = 0
		} else if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	parts = append(parts, n)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid mc version: %s", mcVer)
	}
	minor := parts[1]
	var matched []string
	for _, ver := range v.Versions {
		var vn int
		for _, c := range ver {
			if c == '.' {
				break
			}
			if c >= '0' && c <= '9' {
				vn = vn*10 + int(c-'0')
			}
		}
		if vn == minor {
			matched = append(matched, ver)
		}
	}
	if len(matched) == 0 {
		return "", fmt.Errorf("no neoforge for %d", minor)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(matched)))
	return matched[0], nil
}
