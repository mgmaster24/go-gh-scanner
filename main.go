package main

import (
	"fmt"

	"github.com/mgmaster24/go-gh-scanner/config"
	ghapi "github.com/mgmaster24/go-gh-scanner/github-api"
)

func main() {
	// Get Args for Scan operation

	languages := make(map[string]struct{})
	languages["TypeScript"] = struct{}{}
	languages["JavaScript"] = struct{}{}

	configVals, err := config.GetConfigValues()
	if err != nil {
		fmt.Println(err)
	}

	repos, err := ghapi.GetReposForOrg(configVals["ORG"], configVals["TOKEN"], languages)
	if err != nil {
		fmt.Println(err)
	}

	for _, r := range repos {
		fmt.Printf("Repo - Name: %s, Description: %s, Lang: %s\n", r.Name, r.Description, r.Language)
	}
	// Run scan to get results

	// Apply results to desired infrastructure

}
