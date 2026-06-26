package mcgo

type Rule struct {
	Action   string        `json:"action"` // "allow" or "disallow"
	OS       *RuleOS       `json:"os,omitempty"`
	Features *RuleFeatures `json:"features,omitempty"`
}

type RuleOS struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Arch    string `json:"arch,omitempty"`
}

type RuleFeatures map[string]bool

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

type Download struct {
	Path string `json:"path"`
	URL  string `json:"url"`
	SHA1 string `json:"sha1"`
	Size int64  `json:"size"`
}

// Allowed returns true if a rule set permits this component on the given platform.
func Allowed(rules []Rule, p Platform, features map[string]bool) bool {
	if len(rules) == 0 {
		return true
	}
	allowed := false
	for _, r := range rules {
		if !matches(&r, p, features) {
			continue
		}
		allowed = r.Action == "allow"
	}
	return allowed
}

func matches(r *Rule, p Platform, features map[string]bool) bool {
	if r.OS != nil {
		if r.OS.Name != p.OS {
			return false
		}
		if r.OS.Arch != "" && r.OS.Arch != p.Arch {
			return false
		}
	}
	for feat, need := range *r.Features {
		if need && !features[feat] {
			return false
		}
	}
	return true
}

func (lib *Library) ok(p Platform) bool {
	return Allowed(lib.Rules, p, nil)
}

func (lib *Library) nativeClassifier(p Platform) string {
	if lib.Natives == nil || lib.Downloads == nil || lib.Downloads.Classifiers == nil {
		return ""
	}
	rep := lib.Natives[p.OS]
	if rep == "" {
		return ""
	}
	return rep
}
