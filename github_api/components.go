package github_api

import (
	"fmt"
	"sync"

	"github.com/mgmaster24/go-gh-scanner/models/api_results"
	"github.com/mgmaster24/go-gh-scanner/search"
	"github.com/mgmaster24/go-gh-scanner/utils"
)

type discoveryResult struct {
	tokens []string
	err    error
}

// DiscoverComponentTokens downloads each source repo concurrently, extracts
// component identifiers from its source files, and returns a deduplicated
// token list. Supports Angular (selector values), React (exported PascalCase
// names), and Vue (component name values) source repos.
func (ghClient *GHClient) DiscoverComponentTokens(
	owner string,
	repos []string,
	extractDir string,
	authToken string,
) ([]string, error) {
	results := make([]discoveryResult, len(repos))
	var wg sync.WaitGroup

	for i, repoName := range repos {
		wg.Add(1)
		go func(idx int, name string) {
			defer wg.Done()
			toks, err := ghClient.discoverFromRepo(owner, name, extractDir, authToken)
			results[idx] = discoveryResult{tokens: toks, err: err}
		}(i, repoName)
	}

	wg.Wait()

	seen := make(map[string]bool)
	allTokens := make([]string, 0)
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		for _, t := range r.tokens {
			if !seen[t] {
				seen[t] = true
				allTokens = append(allTokens, t)
			}
		}
	}

	return allTokens, nil
}

func (ghClient *GHClient) discoverFromRepo(owner, repoName, extractDir, authToken string) ([]string, error) {
	fmt.Printf("Discovering components from %s/%s\n", owner, repoName)

	repo, _, err := ghClient.Client.Repositories.Get(ghClient.Ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo %s: %w", repoName, err)
	}

	ghRepo := api_results.GHRepo{
		Name:          *repo.Name,
		APIUrl:        *repo.URL,
		DefaultBranch: *repo.DefaultBranch,
	}

	archiveFile, err := ghRepo.GetRepoArchive(authToken, api_results.Tarball, extractDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get archive for %s: %w", repoName, err)
	}

	dir, err := utils.ExtractGZIP(archiveFile, extractDir)
	utils.RemoveFile(archiveFile)
	if err != nil {
		return nil, err
	}
	defer utils.RemoveDir(dir)

	files, err := utils.GetFilesByExtension(dir, []string{".ts", ".tsx", ".jsx", ".vue"})
	if err != nil {
		return nil, err
	}

	return search.ExtractComponentTokens(files)
}
