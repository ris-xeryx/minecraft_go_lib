package mcgo

type quiltMetaResponse []struct {
	Loader struct {
		Version string `json:"version"`
	} `json:"loader"`
}

func getQuiltVersions(mcVersion string) ([]string, error) {
	url := "https://meta.quiltmc.org/v3/versions/loader/" + mcVersion
	var resp quiltMetaResponse
	if err := httpGetJSON(url, &resp); err != nil {
		return nil, err
	}
	var versions []string
	for _, e := range resp {
		versions = append(versions, e.Loader.Version)
	}
	return versions, nil
}
