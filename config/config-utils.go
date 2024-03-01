// config-utils.go - provide utility methods for handling and manipulating data provided
// in the AppConfig struct instance
package config

import (
	"strings"

	"github.com/google/go-github/github"
	"github.com/mgmaster24/go-gh-scanner/utils"
)

// Returns the Language object associated with the provided string
func (config *AppConfig) GetLanguage(lang string) *Language {
	for _, l := range config.Languages {
		if l.Name == lang {
			return &l
		}
	}

	return nil
}

// Gets a list of file extensions associated with each defined language in the config
//
// Example Language Definition:
//
// Language { Name: "HTML", Extension: ".html" }
func (config *AppConfig) GetLanguageExts() []string {
	exts := make([]string, len(config.Languages))
	for i, l := range config.Languages {
		exts[i] = l.Extension
	}

	return exts
}

// Converts the PerPage ScanConfig value to a githubListOptions pointer
func (config *AppConfig) ToListOptions() *github.ListOptions {
	return &github.ListOptions{PerPage: config.ScanConfig.PerPage}
}

// Gets the short relative value of a dependency
//
// npm deps are generally split by a backslash.  I.E. @angular/material
func (config *AppConfig) GetShortDepName() string {
	depParts := strings.Split(config.CurrentDep, "/")
	return depParts[len(depParts)-1]
}

// Determines whether the repo value is in the ignore repos slice
func (config *AppConfig) ShouldIgnoreRepo(repoName string) bool {
	return isInStrArray(config.ReposToIgnore, repoName)
}

// Determines whether the team value is in the ignore repos slice
func (teamsToIgnore TeamsToIgnore) ShouldIgnoreTeam(teamName string) bool {
	return isInStrArray(teamsToIgnore, teamName)
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
