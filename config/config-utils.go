package config

import (
	"strings"

	"github.com/google/go-github/github"
)

func ReadConfig(fileName string, configVar Config) error {
	return configVar.Read(fileName)
}

func (config *AppConfig) GetLanguagesMap() map[string]struct{} {
	languages := make(map[string]struct{})
	for _, v := range config.Languages {
		languages[v] = struct{}{}
	}

	return languages
}

func (config *AppConfig) ToListOptions() *github.ListOptions {
	return &github.ListOptions{PerPage: config.ScanConfig.PerPage}
}

func (config *AppConfig) GetShortDepName() string {
	// npm deps are generally split by a backslash.  I.E. @angular/material
	depParts := strings.Split(config.CurrentDep, "/")
	return depParts[len(depParts)-1]
}

func (config *AppConfig) ShouldIgnoreRepo(repoName string) bool {
	return isInStrArray(config.ReposToIgnore, repoName)
}

func (config *AppConfig) ShouldIgnoreTeam(teamName string) bool {
	return isInStrArray(config.TeamsToIgnore, teamName)
}

func isInStrArray(vals []string, strToCheck string) bool {
	for _, val := range vals {
		if val == strToCheck {
			return true
		}
	}

	return false
}
