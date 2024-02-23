package config

type AppConfig struct {
	AuthToken  string `json:"authToken"`
	Owner      string `json:"owner"` // can be user or organization
	ExtractDir string `json:"extractDir"`
	DependencyScanConfig
	ScanConfig
	CurrentDep string
}

type ScanConfig struct {
	PerPage int `json:"perPage"`
}

type DependencyScanConfig struct {
	PackageFile   string     `json:"packageFile"`
	Languages     []Language `json:"languages"`
	Dependencies  []string   `json:"dependencies"`
	ReposToIgnore []string   `json:"reposToIgnore"`
	TeamsToIgnore []string   `json:"teamsToIgnore"`
}

type Language struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
}

type Config interface {
	Read(fileName string) error
}
