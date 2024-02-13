package config

import (
	"encoding/json"
	"io"
	"os"
)

func (appConfig *AppConfig) GetConfig() error {
	configFile, err := os.Open(appConfig.Location)
	if err != nil {
		return err
	}

	defer configFile.Close()

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, appConfig)
	if err != nil {
		return err
	}

	return nil
}
