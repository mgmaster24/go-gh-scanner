package api_results

import (
	"time"

	"github.com/mgmaster24/go-gh-scanner/writer"
)

type GHRepo struct {
	Name         string    `json:"repoName"`
	Description  string    `json:"description"`
	Language     string    `json:"language"`
	Owner        string    `json:"owner"`
	Url          string    `json:"url"`
	Team         string    `json:"team"`
	LastModified time.Time `json:"lastModified"`
}

type GHRepoUsingDependencies struct {
	Repo               *GHRepo `json:"repo"`
	DependencyVersions string  `json:"dependencyVersion"`
}

type GHRepoResults struct {
	Repos []GHRepoUsingDependencies `json:"repos"`
	Count int                       `json:"count"`
}

func (ghRepoResults *GHRepoResults) SaveRepoResults(fileName string) error {
	return writer.MarshallAndSave(fileName, ghRepoResults)
}
