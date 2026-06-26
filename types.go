package mcgo

// ── Platform ─────────────────────────────────────────────────────────────

type Platform struct {
	OS   string // "windows", "osx", "linux"
	Arch string // "x86", "x64", "arm64"
}

func Host() Platform { return Platform{OS: currentOS(), Arch: currentArch()} }

// ── Mojang API types ─────────────────────────────────────────────────────

type ManifestVersion struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	URL     string `json:"url"`
	Release string `json:"time"`
}

type VersionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []ManifestVersion `json:"versions"`
}

type VersionInfo struct {
	Arguments     *VersionArguments `json:"arguments,omitempty"`
	GameArguments []string          `json:"minecraftArguments,omitempty"`
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
	SHA1      string `json:"sha1"`
	TotalSize int64  `json:"totalSize"`
}

type VersionDownloads struct {
	Client *Download `json:"client"`
	Server *Download `json:"server"`
}

type JavaVersion struct {
	Component    string `json:"component"`
	MajorVersion int    `json:"majorVersion"`
}

// ── Profile ──────────────────────────────────────────────────────────────

type Profile struct {
	Username    string `json:"username"`
	UUID        string `json:"uuid"`
	AccessToken string `json:"access_token,omitempty"`
}

// ── Loader ───────────────────────────────────────────────────────────────

type Loader string

const (
	Vanilla  Loader = "Vanilla"
	Fabric   Loader = "Fabric"
	Quilt    Loader = "Quilt"
	Forge    Loader = "Forge"
	NeoForge Loader = "NeoForge"
)

func Loaders() []Loader {
	return []Loader{Vanilla, Fabric, Quilt, NeoForge, Forge}
}

// ── Launch ───────────────────────────────────────────────────────────────

type LaunchOpts struct {
	Version string
	Loader  Loader
	Dir     string // where .minecraft directory lives
	Profile Profile
	MinRAM  string // "2G"
	MaxRAM  string // "4G"
	JVMArgs []string
	Bus     *EventBus
}
