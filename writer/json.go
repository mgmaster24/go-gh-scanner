package writer

import (
	"encoding/json"
	"os"

	api_results "github.com/mgmaster24/go-gh-scanner/models/api-results"
)

func SaveScanResults(fileName string, results api_results.ScanResults) error {
	return marshallAndSave(fileName, results)
}

func SaveRepoResults(fileName string, repoResults api_results.GHRepoResults) error {
	return marshallAndSave(fileName, repoResults)
}

func marshallAndSave(fileName string, val any) error {
	json, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = os.WriteFile(fileName, json, 0644)
	return err
}
