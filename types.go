package mcgo

// Platform identifica el sistema operativo + arquitectura.
type Platform struct {
	OS   string // "windows", "osx" (macOS), "linux"
	Arch string // "x86", "x64", "arm64"
}

// CurrentPlatform detecta la plataforma actual.
func CurrentPlatform() Platform {
	return Platform{
		OS:   currentOSName(),
		Arch: currentArchName(),
	}
}

// Version es una versión disponible del manifiesto de Mojang.
type Version struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	URL     string `json:"url"`
	Release string `json:"time"`
}

// VersionManifest es la respuesta de version_manifest_v2.json
type VersionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []Version `json:"versions"`
}

// Profile es el resultado de autenticación (offline o Microsoft).
type Profile struct {
	Username    string `json:"username"`
	UUID        string `json:"uuid"`
	AccessToken string `json:"access_token,omitempty"`
	PlayerName  string `json:"player_name,omitempty"`
	XUID        string `json:"xuid,omitempty"`
}

// LoaderType identifica el loader (Vanilla, Fabric, Quilt, Forge, NeoForge).
type LoaderType string

const (
	LoaderVanilla  LoaderType = "Vanilla"
	LoaderFabric   LoaderType = "Fabric"
	LoaderQuilt    LoaderType = "Quilt"
	LoaderForge    LoaderType = "Forge"
	LoaderNeoForge LoaderType = "NeoForge"
)

// Instance es una instalación de Minecraft: versión + loader.
type Instance struct {
	Name          string
	Version       string
	Loader        LoaderType
	LoaderVersion string
	Directory     string
}

// LaunchMemory define la memoria RAM del proceso.
type LaunchMemory struct {
	Min string // "2G"
	Max string // "4G"
}

// LaunchOptions son las opciones para lanzar Minecraft.
type LaunchOptions struct {
	Instance    Instance
	Profile     Profile
	Memory      LaunchMemory
	JVMArgs     []string
	GameArgs    []string
	JavaPath    string
	NativesPath string
	EventBus    *EventBus
}
