package mcgo

import (
	"fmt"
	"path/filepath"
)

// Install downloads a Minecraft version to the given directory.
// Returns the parsed version info.
func Install(version, dir string, bus *EventBus) (*VersionInfo, error) {
	if bus == nil {
		bus = NewEventBus()
	}
	info, err := fetchVersion(version)
	if err != nil {
		return nil, err
	}
	p := Host()
	if info.InheritsFrom != "" {
		if parent, err := fetchVersion(info.InheritsFrom); err == nil {
			merge(info, parent)
		}
	}
	bus.emitInstallStarted(0)
	if err := installClient(info, dir, bus); err != nil {
		return info, err
	}
	if err := installLibs(info, dir, p, bus); err != nil {
		return info, err
	}
	if err := installAssets(info, dir, bus); err != nil {
		return info, err
	}
	if err := installNatives(info, dir, p, bus); err != nil {
		return info, err
	}
	bus.emitInstallComplete()
	return info, nil
}

// fetchVersion gets the detailed JSON for a version.
func fetchVersion(id string) (*VersionInfo, error) {
	url, err := versionURL(id)
	if err != nil {
		return nil, err
	}
	var vi VersionInfo
	if err := fetchJSON(url, &vi); err != nil {
		return nil, fmt.Errorf("version %s: %w", id, err)
	}
	return &vi, nil
}

func installClient(info *VersionInfo, dir string, bus *EventBus) error {
	if info.Downloads == nil || info.Downloads.Client == nil {
		return nil
	}
	dl := info.Downloads.Client
	path := filepath.Join(dir, "versions", info.ID, info.ID+".jar")
	return downloadIfMissing(dl.URL, path, dl.SHA1, bus)
}

func installLibs(info *VersionInfo, dir string, p Platform, bus *EventBus) error {
	base := filepath.Join(dir, "libraries")
	for i := range info.Libraries {
		lib := &info.Libraries[i]
		if !lib.ok(p) || lib.Downloads == nil || lib.Downloads.Artifact == nil {
			continue
		}
		dl := lib.Downloads.Artifact
		path := filepath.Join(base, dl.Path)
		if err := downloadIfMissing(dl.URL, path, dl.SHA1, bus); err != nil {
			return fmt.Errorf("%s: %w", lib.Name, err)
		}
	}
	return nil
}

func installAssets(info *VersionInfo, dir string, bus *EventBus) error {
	if info.AssetIndex == nil {
		return nil
	}
	ai := info.AssetIndex
	idxPath := filepath.Join(dir, "assets", "indexes", ai.ID+".json")
	if err := downloadIfMissing(ai.URL, idxPath, ai.SHA1, bus); err != nil {
		return err
	}
	var idx struct {
		Objects map[string]struct {
			Hash string `json:"hash"`
			Size int64  `json:"size"`
		} `json:"objects"`
	}
	if err := readJSON(idxPath, &idx); err != nil {
		return err
	}
	base := filepath.Join(dir, "assets", "objects")
	for _, obj := range idx.Objects {
		h := obj.Hash
		url := "https://resources.download.minecraft.net/" + h[:2] + "/" + h
		path := filepath.Join(base, h[:2], h)
		if err := downloadIfMissing(url, path, h, bus); err != nil {
			return err
		}
	}
	return nil
}

func installNatives(info *VersionInfo, dir string, p Platform, bus *EventBus) error {
	out := filepath.Join(dir, "natives", info.ID)
	if err := mkdir(out); err != nil {
		return err
	}
	base := filepath.Join(dir, "libraries")
	for i := range info.Libraries {
		lib := &info.Libraries[i]
		if !lib.ok(p) {
			continue
		}
		name := lib.nativeClassifier(p)
		if name == "" || lib.Downloads == nil || lib.Downloads.Classifiers == nil {
			continue
		}
		dl := lib.Downloads.Classifiers[name]
		if dl == nil {
			continue
		}
		path := filepath.Join(base, dl.Path)
		if err := downloadIfMissing(dl.URL, path, dl.SHA1, bus); err != nil {
			return err
		}
	}
	return nil
}

// merge fills child fields from parent when empty (for inheritsFrom versions).
func merge(child, parent *VersionInfo) {
	if child.MainClass == "" {
		child.MainClass = parent.MainClass
	}
	if child.Assets == "" {
		child.Assets = parent.Assets
	}
	if child.AssetIndex == nil {
		child.AssetIndex = parent.AssetIndex
	}
	if child.Downloads == nil {
		child.Downloads = parent.Downloads
	}
	if child.JavaVersion == nil {
		child.JavaVersion = parent.JavaVersion
	}
	if child.GameArguments == nil {
		child.GameArguments = parent.GameArguments
	}
	if child.Arguments == nil {
		child.Arguments = parent.Arguments
	}
	if parent.JAR != "" && child.ID == "" {
		child.ID = parent.ID
	}
	if parent.JAR != "" && child.JAR == "" {
		child.JAR = parent.JAR
	}
	seen := map[string]*Library{}
	for i := range parent.Libraries {
		seen[parent.Libraries[i].Name] = &parent.Libraries[i]
	}
	for i := range child.Libraries {
		seen[child.Libraries[i].Name] = &child.Libraries[i]
	}
	child.Libraries = nil
	for _, lib := range seen {
		child.Libraries = append(child.Libraries, *lib)
	}
}
