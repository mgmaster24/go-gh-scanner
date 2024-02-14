package config

import (
	"github.com/mgmaster24/go-gh-scanner/reader"
)

func (appConfig *AppConfig) Read(fileName string) error {
	config, err := reader.ReadJSONData[AppConfig](fileName)
	if err != nil {
		return err
	}

	*appConfig = config

	return nil
}
