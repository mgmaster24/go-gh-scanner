package config

import (
	"errors"
	"os"
	"strings"
)

type EnvCongifReader struct{}

func (configReader *EnvCongifReader) GetConfigValues() (*AppConfig, error) {
	appConfig := &AppConfig{}
	val := os.Getenv("ORG")
	if val == "" {
		return nil, errors.New("couldn't retrieve envoriment variable")
	}

	appConfig.Organization = val

	val = os.Getenv("TOKEN")
	if val == "" {
		return nil, errors.New("couldn't retrieve envoriment variable")
	}
	appConfig.GHAuthToken = val
	appConfig.Languages = splitVar("Languages")
	appConfig.Dependencies = splitVar("Dependencieds")
	appConfig.SearchStrings = splitVar("SearchStrings")

	return appConfig, nil
}

func splitVar(strVar string) []string {
	val := os.Getenv("SearchStrings")
	ss := strings.Split(val, " ")
	return ss
}
