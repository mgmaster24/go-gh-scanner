package config

import (
	"testing"
)

func TestAppConfigReader(t *testing.T) {
	appConfig, err := Read("app-config-test.json")
	if err != nil {
		t.Fatal(err)
	}

	verifySettings(appConfig, t)
}

func verifySettings(appConfig *AppConfig, t *testing.T) {
	if appConfig.PerPage != 100 {
		t.Fatal("Failed to read per page")
	}

	if appConfig.AuthTokenKey != "secret-gh-token-key" {
		t.Fatal("Failed to read auth token")
	}

	if len(appConfig.Languages) == 0 {
		t.Fatal("Failed to read languages")
	}

	if len(appConfig.Languages) != 3 {
		t.Fatal("Wrong number of languages")
	}

	for _, lang := range appConfig.Languages {
		if lang.Name != "TypeScript" && lang.Name != "JavaScript" && lang.Name != "HTML" {
			t.Fatal("Failed to read languages")
		}
	}

	exts := appConfig.GetLanguageExts()
	if len(exts) != 3 {
		t.Fatal("Failed to get the correct number of extensions")
	}

	if appConfig.PackageFile != "package.json" {
		t.Fatal("Failed to read package file")
	}

	if len(appConfig.ReposToIgnore) == 0 {
		t.Fatal("Failed to read ReposToIgnore")
	}

	// Should be one repo in config
	if len(appConfig.ReposToIgnore) > 1 {
		t.Fatal("Read incorrect number of repos to ignore")
	}

	if appConfig.ReposToIgnore[0] != "ignore-this-repo" {
		t.Fatal("Incorrect repo read")
	}

	if appConfig.ResultsConfig.Destination == "" {
		t.Fatal("Failed to read results writer config")
	}

	if appConfig.ResultsConfig.DestinationType == "" {
		t.Fatal("Failed to read results writer config destination type")
	}

	if appConfig.ResultsConfig.Destination != "my-scanner-table" {
		t.Fatal("Failed to read results writer config destination")
	}

	if appConfig.ResultsConfig.DestinationType != TableDestination {
		t.Fatal("Failed to read results writer config destination type")
	}

	if appConfig.ResultsConfig.UseBatchProcessing != true {
		t.Fatal("Failed to read results writer config batch processing value")
	}
}

func TestShouldContainRepo(t *testing.T) {
	reposToIgnore := []string{"test-repo", "test2-repo", "*-this-is-a-wildcard", "test-*-wildcard"}
	appConfig := AppConfig{}
	appConfig.ReposToIgnore = reposToIgnore

	if !appConfig.ShouldIgnoreRepo("test-repo") {
		t.Fatal("Failed to ignore expected repo")
	}

	if !appConfig.ShouldIgnoreRepo("test2-repo") {
		t.Fatal("Failed to ignore expected repo")
	}

	if !appConfig.ShouldIgnoreRepo("should-ignore-this-is-a-wildcard") {
		t.Fatal("Failed to ignore expected repo")
	}

	if !appConfig.ShouldIgnoreRepo("test-this-is-a-wildcard") {
		t.Fatal("Failed to ignore expected repo")
	}
}
