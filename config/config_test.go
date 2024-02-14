package config

import (
	"testing"
)

func TestAppConfigReader(t *testing.T) {
	appConfig := &AppConfig{}
	err := appConfig.Read("app-config-test.json")
	if err != nil {
		t.Fatal(err)
	}

	verifySettings(appConfig, t)
}

func TestAppConfigReaderUtil(t *testing.T) {
	appConfig := &AppConfig{}
	err := ReadConfig("app-config-test.json", appConfig)
	if err != nil {
		t.Fatal(err)
	}

	verifySettings(appConfig, t)
}

func verifySettings(appConfig *AppConfig, t *testing.T) {
	if appConfig.PerPage != 100 {
		t.Fatal("Failed to read per page")
	}

	if appConfig.GHAuthToken != "secret-gh-token" {
		t.Fatal("Failed to read auth token")
	}

	if appConfig.Organization != "MyOrganization" {
		t.Fatal("Failed to read organization")
	}

	if len(appConfig.Languages) == 0 {
		t.Fatal("Failed to read languages")
	}

	for _, lang := range appConfig.Languages {
		if lang != "TypeScript" && lang != "JavaScript" {
			t.Fatal("Failed to read languages")
		}
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

	if len(appConfig.TeamsToIgnore) == 0 {
		t.Fatal("Failed to read ReposToIgnore")
	}

	if len(appConfig.TeamsToIgnore) > 4 {
		t.Fatal("Read incorrect number of teams to ignore")
	}

	if appConfig.TeamsToIgnore[0] != "software-team" {
		t.Fatal("Incorrect team read. Expected software-team")
	}

	if appConfig.TeamsToIgnore[1] != "pd-team" {
		t.Fatal("Incorrect team read. Expected pd-team")
	}

	if appConfig.TeamsToIgnore[2] != "qa-team" {
		t.Fatal("Incorrect team read. Expected qa-team")
	}

	if appConfig.TeamsToIgnore[3] != "security-team" {
		t.Fatal("Incorrect team read. Expected security-team")
	}
}
