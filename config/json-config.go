package config

import (
	"encoding/json"
	"io"
	"os"
)

type JSONConfigReader struct {
	ScanConfigFile string
}

func (jsonReader *JSONConfigReader) GetConfigValues() (*AppConfig, error) {
	configFile, err := os.Open(jsonReader.ScanConfigFile)
	if err != nil {
		return nil, err
	}

	defer configFile.Close()

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		return nil, err
	}

	appConfig := &AppConfig{}
	err = json.Unmarshal(bytes, appConfig)
	if err != nil {
		return nil, err
	}

	return appConfig, nil
}
