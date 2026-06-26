package mcgo

import (
	"fmt"
	"os"
	"path/filepath"
)

// VersionInfo es la respuesta del JSON de una versión específica.
type VersionInfo struct {
	Arguments     *VersionArguments `json:"arguments,omitempty"`
	GameArguments []string          `json:"minecraftArguments,omitempty"` // legacy
	Libraries     []Library         `json:"libraries"`
	AssetIndex    *AssetIndex       `json:"assetIndex,omitempty"`
	Assets        string            `json:"assets,omitempty"`
	Downloads     *VersionDownloads `json:"downloads"`
	JavaVersion   *JavaVersion      `json:"javaVersion,omitempty"`
	MainClass     string            `json:"mainClass"`
	ID            string            `json:"id"`
	Type          string            `json:"type"`
	InheritsFrom  string            `json:"inheritsFrom,omitempty"`
	JAR           string            `json:"jar,omitempty"`
}

type VersionArguments struct {
	Game []VersionArg `json:"game"`
	JVM  []VersionArg `json:"jvm"`
}

type VersionArg struct {
	Value []string `json:"value,omitempty"`
	Rules []Rule   `json:"rules,omitempty"`
}

type AssetIndex struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	TotalSize int64  `json:"totalSize"`
	SHA1      string `json:"sha1"`
	Size      int64  `json:"size"`
}

type VersionDownloads struct {
	Client *Download `json:"client"`
	Server *Download `json:"server"`
}

type JavaVersion struct {
	Component    string `json:"component"`
	MajorVersion int    `json:"majorVersion"`
}

// GetVersionInfo obtiene el JSON de una versión específica.
func GetVersionInfo(versionID string) (*VersionInfo, error) {
	url, err := GetVersionURL(versionID)
	if err != nil {
		return nil, err
	}
	var vi VersionInfo
	if err := httpGetJSON(url, &vi); err != nil {
		return nil, fmt.Errorf("fetching version %s: %w", versionID, err)
	}
	return &vi, nil
}

// InstallParams configura la instalación.
type InstallParams struct {
	Version   string
	Directory string
	Platform  Platform
	EventBus  *EventBus
}

// Install instala una versión de Minecraft (client, libraries, assets, natives).
func Install(params InstallParams) (*VersionInfo, error) {
	if params.Platform.OS == "" {
		params.Platform = CurrentPlatform()
	}
	if params.EventBus == nil {
		params.EventBus = NewEventBus()
	}

	info, err := GetVersionInfo(params.Version)
	if err != nil {
		return nil, err
	}

	// Heredar de versión padre si aplica
	if info.InheritsFrom != "" {
		parent, err := GetVersionInfo(info.InheritsFrom)
		if err != nil {
			return nil, err
		}
		mergeVersions(info, parent)
	}

	bus := params.EventBus
	bus.emitInstallStarted(0)

	if err := installClient(info, params, bus); err != nil {
		return info, err
	}
	if err := installLibraries(info, params, bus); err != nil {
		return info, err
	}
	if err := installAssets(info, params, bus); err != nil {
		return info, err
	}
	if err := installNatives(info, params, bus); err != nil {
		return info, err
	}

	bus.emitInstallComplete()
	return info, nil
}

// installClient — descargar el client.jar
func installClient(info *VersionInfo, params InstallParams, bus *EventBus) error {
	if info.Downloads == nil || info.Downloads.Client == nil {
		return nil
	}
	dl := info.Downloads.Client
	path := filepath.Join(params.Directory, "versions", info.ID, info.ID+".jar")
	return downloadIfMissing(dl.URL, path, dl.SHA1, bus)
}

// installLibraries — descargar todas las librerías que aplican
func installLibraries(info *VersionInfo, params InstallParams, bus *EventBus) error {
	libsDir := filepath.Join(params.Directory, "libraries")
	for i := range info.Libraries {
		lib := &info.Libraries[i]
		if !lib.AllowedFor(params.Platform) {
			continue
		}
		if lib.Downloads == nil || lib.Downloads.Artifact == nil {
			continue
		}
		dl := lib.Downloads.Artifact
		path := filepath.Join(libsDir, dl.Path)
		if err := downloadIfMissing(dl.URL, path, dl.SHA1, bus); err != nil {
			return fmt.Errorf("library %s: %w", lib.Name, err)
		}
	}
	return nil
}

// installAssets — descargar el asset index y todos los objetos
func installAssets(info *VersionInfo, params InstallParams, bus *EventBus) error {
	if info.AssetIndex == nil {
		return nil
	}
	ai := info.AssetIndex

	// Descargar asset index
	indexPath := filepath.Join(params.Directory, "assets", "indexes", ai.ID+".json")
	if err := downloadIfMissing(ai.URL, indexPath, ai.SHA1, bus); err != nil {
		return fmt.Errorf("asset index: %w", err)
	}

	// Parsear el index
	var idx struct {
		Objects map[string]struct {
			Hash string `json:"hash"`
			Size int64  `json:"size"`
		} `json:"objects"`
	}
	if err := readJSONLocal(indexPath, &idx); err != nil {
		return err
	}

	objectsDir := filepath.Join(params.Directory, "assets", "objects")
	for _, obj := range idx.Objects {
		url := "https://resources.download.minecraft.net/" + obj.Hash[:2] + "/" + obj.Hash
		path := filepath.Join(objectsDir, obj.Hash[:2], obj.Hash)
		if err := downloadIfMissing(url, path, obj.Hash, bus); err != nil {
			return err
		}
	}
	return nil
}

// installNatives — descargar y extraer natives
func installNatives(info *VersionInfo, params InstallParams, bus *EventBus) error {
	nativesDir := filepath.Join(params.Directory, "natives", info.ID)
	if err := ensureDir(nativesDir); err != nil {
		return err
	}

	libsDir := filepath.Join(params.Directory, "libraries")
	for i := range info.Libraries {
		lib := &info.Libraries[i]
		if !lib.AllowedFor(params.Platform) {
			continue
		}
		nativeName := lib.NativeClassifier(params.Platform)
		if nativeName == "" {
			continue
		}
		if lib.Downloads == nil || lib.Downloads.Classifiers == nil {
			continue
		}
		dl := lib.Downloads.Classifiers[nativeName]
		if dl == nil {
			continue
		}
		path := filepath.Join(libsDir, dl.Path)
		if err := downloadIfMissing(dl.URL, path, dl.SHA1, bus); err != nil {
			return err
		}
		// TODO: extraer .so/.dll/.dylib del zip (omitido en MVP)
		_ = nativesDir
	}
	return nil
}

// mergeVersions — mezclar version info hija con padre (inheritsFrom)
func mergeVersions(child, parent *VersionInfo) {
	if child.MainClass == "" {
		child.MainClass = parent.MainClass
	}
	if child.Assets == "" {
		child.Assets = parent.Assets
	}
	if child.AssetIndex == nil && parent.AssetIndex != nil {
		child.AssetIndex = parent.AssetIndex
	}
	if child.Downloads == nil && parent.Downloads != nil {
		child.Downloads = parent.Downloads
	}
	if child.JavaVersion == nil && parent.JavaVersion != nil {
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
	// Mezclar libraries (padre primero, reemplazar si mismo name)
	libMap := make(map[string]*Library)
	for i := range parent.Libraries {
		libMap[parent.Libraries[i].Name] = &parent.Libraries[i]
	}
	for i := range child.Libraries {
		libMap[child.Libraries[i].Name] = &child.Libraries[i]
	}
	child.Libraries = nil
	for _, lib := range libMap {
		child.Libraries = append(child.Libraries, *lib)
	}
}

// readJSONLocal parsea JSON desde un archivo local.
func readJSONLocal(path string, v interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return jsonDecode(f, v)
}
