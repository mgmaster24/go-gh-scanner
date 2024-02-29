package config

type AppConfig struct {
	AuthTokenKey string `json:"authTokenKey"`
	Owner        string `json:"owner"` // can be user or organization
	ExtractDir   string `json:"extractDir"`
	WriterConfig `json:"writerConfig"`
	SecretKeys   `json:"secretKeys"`
	DependencyScanConfig
	ScanConfig
	CloudConfig

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
	TableDesitnation   DestinationType = "table"
)

type Config interface {
	Read(fileName string) error
}
