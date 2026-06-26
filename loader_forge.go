package mcgo

import "sort"

// Forge promotions: <MC>-recommended o <MC>-latest
type forgePromotionsResponse struct {
	Promos map[string]string `json:"promos"`
}

func getForgeVersions(mcVersion string) ([]string, error) {
	url := "https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json"
	var resp forgePromotionsResponse
	if err := httpGetJSON(url, &resp); err != nil {
		return nil, err
	}
	var versions []string
	// recommended primero, luego latest
	if v, ok := resp.Promos[mcVersion+"-recommended"]; ok {
		versions = append(versions, v)
	}
	if v, ok := resp.Promos[mcVersion+"-latest"]; ok {
		// evitar duplicar
		if len(versions) == 0 || versions[0] != v {
			versions = append(versions, v)
		}
	}
	// Buscar también todas las versiones que empiecen con el MC version
	// (esto captura casos donde solo hay un "-latest")
	for key, v := range resp.Promos {
		if key == mcVersion+"-"+v || key == mcVersion+"-latest" || key == mcVersion+"-recommended" {
			continue
		}
	}
	return versions, nil
}

// NeoForge: <minor>-<patch>, donde minor es el segundo número de MC version (1.21 -> 21)
type neoforgeVersionsResponse struct {
	Versions []string `json:"versions"`
}

func getNeoForgeVersions(mcVersion string) ([]string, error) {
	url := "https://maven.neoforged.net/api/maven/versions/releases/net/neoforged/neoforge"
	var resp neoforgeVersionsResponse
	if err := httpGetJSON(url, &resp); err != nil {
		return nil, err
	}
	// MC 1.21 -> buscar versiones que empiecen con "21.0.0" etc.
	// NeoForge usa el minorMC como major: 1.21.x -> 21.x.y
	var parts []int
	for _, p := range splitDots(mcVersion) {
		n := atoiSafe(p)
		parts = append(parts, n)
	}
	if len(parts) < 2 {
		return nil, nil
	}
	minor := parts[1]
	var matched []string
	for _, v := range resp.Versions {
		vParts := splitDots(v)
		vFirst := atoiSafe(vParts[0])
		if vFirst == minor {
			matched = append(matched, v)
		}
	}
	// Ordenar descendente
	sort.Sort(sort.Reverse(sort.StringSlice(matched)))
	return matched, nil
}

func splitDots(s string) []string {
	var out []string
	start := 0
	for i, c := range s {
		if c == '.' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

func atoiSafe(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}
