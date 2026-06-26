package mcgo

// Rule es una regla de Mojang para decidir si incluir una library/native.
type Rule struct {
	Action   string        `json:"action"` // "allow" o "disallow"
	OS       *RuleOS       `json:"os,omitempty"`
	Features *RuleFeatures `json:"features,omitempty"`
}

type RuleOS struct {
	Name    string `json:"name"` // "windows", "osx", "linux"
	Version string `json:"version,omitempty"`
	Arch    string `json:"arch,omitempty"`
}

type RuleFeatures map[string]bool

// RuleAllowed evalúa si una lista de reglas permite el archivo en la plataforma actual.
func RuleAllowed(rules []Rule, platform Platform, features map[string]bool) bool {
	if len(rules) == 0 {
		return true
	}
	allowed := false
	for _, r := range rules {
		if !ruleMatches(&r, platform, features) {
			continue
		}
		allowed = r.Action == "allow"
	}
	return allowed
}

func ruleMatches(r *Rule, platform Platform, features map[string]bool) bool {
	if r.OS != nil {
		if r.OS.Name != platform.OS {
			return false
		}
		if r.OS.Arch != "" && r.OS.Arch != platform.Arch {
			return false
		}
	}
	if r.Features != nil {
		for feat, needed := range *r.Features {
			hasFeature := features[feat]
			if needed && !hasFeature {
				return false
			}
		}
	}
	return true
}

// LibraryRefers-a-la-definición-de-librería de Mojang.
type Library struct {
	Name      string            `json:"name"`
	Downloads *LibraryDownloads `json:"downloads,omitempty"`
	Natives   map[string]string `json:"natives,omitempty"`
	Rules     []Rule            `json:"rules,omitempty"`
}

type LibraryDownloads struct {
	Artifact    *Download            `json:"artifact,omitempty"`
	Classifiers map[string]*Download `json:"classifiers,omitempty"`
}

// Download es un archivo descargable con SHA1.
type Download struct {
	Path string `json:"path"`
	URL  string `json:"url"`
	SHA1 string `json:"sha1"`
	Size int64  `json:"size"`
}

// NativesFor 当前 plataforma actual.
func (l *Library) NativesFor(platform Platform) string {
	if l.Natives == nil {
		return ""
	}
	return l.Natives[platform.OS]
}

// AllowedFor indica si la librería aplica a la plataforma actual.
func (l *Library) AllowedFor(platform Platform) bool {
	return RuleAllowed(l.Rules, platform, nil)
}

// NativeKW native classifiers para la plataforma actual.
func (l *Library) NativeClassifier(platform Platform) string {
	if l.Natives == nil {
		return ""
	}
	nativeName := l.Natives[platform.OS]
	if nativeName == "" {
		return ""
	}
	if l.Downloads == nil || l.Downloads.Classifiers == nil {
		return ""
	}
	return nativeName
}
