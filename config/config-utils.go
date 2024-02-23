package config

import (
	"strings"

	"github.com/google/go-github/github"
	"github.com/mgmaster24/go-gh-scanner/utils"
)

func ReadConfig(fileName string, configVar Config) error {
	return configVar.Read(fileName)
}

func (config *AppConfig) GetLanguage(lang string) *Language {
	for _, l := range config.Languages {
		if l.Name == lang {
			return &l
		}
	}

	return nil
}

func (config *AppConfig) GetLanguageExts() []string {
	exts := make([]string, len(config.Languages))
	for i, l := range config.Languages {
		exts[i] = l.Extension
	}

	return exts
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
		if strings.Contains(val, "*") {
			if utils.StringsMatch(val, strToCheck) {
				return true
			}
		} else if val == strToCheck {
			return true
		}
	}

	return false
}
