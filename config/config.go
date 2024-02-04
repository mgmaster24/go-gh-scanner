package config

type Config interface {
	GetConfigValues() map[string]string
}
