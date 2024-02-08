package config

import "strings"

type AppConfig struct {
	GHAuthToken   string   `json:"ghToken"`
	Organization  string   `json:"organization"`
	PackageFile   string   `json:"packageFile"`
	Languages     []string `json:"languages"`
	Dependencies  []string `json:"dependencies"`
	SearchStrings []string `json:"searchStrings"`
	ReposToIgnore []string `json:"reposToIgnore"`
	TeamsToIgnore []string `json:"teamsToIgnore"`
	PerPage       int      `json:"perPage"`
	CurrentDep    string
}

type ConfigReader interface {
	GetConfigValues() (AppConfig, error)
}

func (config *AppConfig) GetLanguagesMap() map[string]struct{} {
	languages := make(map[string]struct{})
	for _, v := range config.Languages {
		languages[v] = struct{}{}
	}

	return languages
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
