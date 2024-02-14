package config

import (
	"errors"
	"os"
	"strings"
)

func GetConfigValues() (*AppConfig, error) {
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

	return appConfig, nil
}

func splitVar(strVar string) []string {
	val := os.Getenv(strVar)
	ss := strings.Split(val, " ")
	return ss
}
