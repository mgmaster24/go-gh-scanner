package config

import (
	"github.com/mgmaster24/go-gh-scanner/reader"
)

func Read(fileName string) (*AppConfig, error) {
	config, err := reader.ReadJSONData[AppConfig](fileName)
	if err != nil {
		return nil, err
	}

	return &config, err
}
