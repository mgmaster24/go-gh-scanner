package github_api

import (
	"fmt"
	"path/filepath"
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

// DiscoverComponentTokensFromMonorepo downloads a single monorepo and extracts
// component tokens from each workspace path. If paths is empty the entire repo is scanned.
func (ghClient *GHClient) DiscoverComponentTokensFromMonorepo(
	owner, repoName string,
	paths []string,
	extractDir, authToken string,
) ([]string, error) {
	fmt.Printf("Discovering components from monorepo %s/%s\n", owner, repoName)

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

	scanDirs := []string{dir}
	if len(paths) > 0 {
		scanDirs = make([]string, len(paths))
		for i, p := range paths {
			scanDirs[i] = filepath.Join(dir, p)
		}
	}

	var allFiles []string
	for _, d := range scanDirs {
		files, err := utils.GetFilesByExtension(d, []string{".ts", ".tsx", ".jsx", ".vue"})
		if err != nil {
			return nil, err
		}
		allFiles = append(allFiles, files...)
	}

	return search.ExtractComponentTokens(allFiles)
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
