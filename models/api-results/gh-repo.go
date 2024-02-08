package api_results

import "time"

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
