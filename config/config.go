package config

type AppConfig struct {
	AuthTokenKey       string                  `json:"authTokenKey"`
	Owner              string                  `json:"owner"` // optional — omit to scan all public GitHub repos
	ExtractDir         string                  `json:"extractDir"`
	ResultsConfig      WriterConfig            `json:"resultsWriterConfig"`
	ComponentDiscovery ComponentDiscoveryConfig `json:"componentDiscovery"`
	SecretKeys         `json:"secretKeys"`
	DependencyScanConfig
	ScanConfig
	CloudConfig

	CurrentDep string
}

// ComponentDiscoveryConfig points at the m2s2 source repos to scan for component definitions.
type ComponentDiscoveryConfig struct {
	Owner string   `json:"owner"`
	Repos []string `json:"repos"`
}

type ScanConfig struct {
	PerPage int `json:"perPage"`
}

type DependencyScanConfig struct {
	PackageFile   string     `json:"packageFile"`
	Languages     []Language `json:"languages"`
	Dependencies  []string   `json:"dependencies"`
	ReposToIgnore []string   `json:"reposToIgnore"`
}

type Language struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
}

type CloudConfig struct {
	Region string `json:"region"`
	DBConfig
}

type DBConfig struct {
	TableName string `json:"tableName"`
}

// Configuration that is used to define a results writer
type WriterConfig struct {
	// Out Directory, table name, etc...
	DestinationType    DestinationType `json:"destinationType"`
	Destination        string          `json:"destination"`
	UseBatchProcessing bool            `json:"useBatch"`
}

// Custom type for defining destinations
type DestinationType string

// SecretKeys
type SecretKeys []string

// Destination type enumeration
//
// Where the results will be written to.
const (
	DefaultDestination DestinationType = "none"
	FileDestination    DestinationType = "file"
	TableDestination   DestinationType = "table"
)

type Config interface {
	Read(fileName string) error
}
