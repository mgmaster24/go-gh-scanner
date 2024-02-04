package config

import (
	"errors"
	"os"
)

func GetConfigValues() (map[string]string, error) {
	configVals := make(map[string]string)

	val := os.Getenv("ORG")
	if val == "" {
		return nil, errors.New("Couldn't retrieve envoriment variable!")
	}

	configVals["ORG"] = val

	val = os.Getenv("TOKEN")
	if val == "" {
		return nil, errors.New("Couldn't retrieve envoriment variable!")
	}

	configVals["TOKEN"] = val

	return configVals, nil
}
