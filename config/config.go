package config

type AppConfig struct {
	// Location can be a file, api, etc.
	Location             string
	GHAuthToken          string `json:"ghToken"`
	Organization         string `json:"organization"`
	DependencyScanConfig `json:"depsConfig"`
	CurrentDep           string
}

type ScanConfig struct {
	PerPage int `json:"perPage"`
}
type DependencyScanConfig struct {
	PackageFile   string   `json:"packageFile"`
	Languages     []string `json:"languages"`
	Dependencies  []string `json:"dependencies"`
	SearchStrings []string `json:"searchStrings"`
	ReposToIgnore []string `json:"reposToIgnore"`
	TeamsToIgnore []string `json:"teamsToIgnore"`
	ScanConfig    ScanConfig
}

type Config interface {
	GetConfig() error
}
