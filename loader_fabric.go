package mcgo

type fabricMetaResponse []struct {
	Loader struct {
		Version string `json:"version"`
	} `json:"loader"`
}

func getFabricVersions(mcVersion string) ([]string, error) {
	url := "https://meta.fabricmc.net/v2/versions/loader/" + mcVersion
	var resp fabricMetaResponse
	if err := httpGetJSON(url, &resp); err != nil {
		return nil, err
	}
	var versions []string
	for _, e := range resp {
		versions = append(versions, e.Loader.Version)
	}
	return versions, nil
}
