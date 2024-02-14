package config

type AppConfig struct {
	GHAuthToken  string `json:"ghToken"`
	Organization string `json:"organization"`
	DependencyScanConfig
	ScanConfig
	CurrentDep string
}

type ScanConfig struct {
	PerPage int `json:"perPage"`
}

type DependencyScanConfig struct {
	PackageFile   string   `json:"packageFile"`
	Languages     []string `json:"languages"`
	Dependencies  []string `json:"dependencies"`
	ReposToIgnore []string `json:"reposToIgnore"`
	TeamsToIgnore []string `json:"teamsToIgnore"`
}

type Config interface {
	Read(fileName string) error
}
