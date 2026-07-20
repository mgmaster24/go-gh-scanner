package search

import (
	"fmt"
	"sync"
	"time"

	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
	"github.com/mgmaster24/go-gh-scanner/utils"
	"github.com/mgmaster24/go-gh-scanner/writer/results"
)

// SearchAndWriteResultsAsync runs SearchAndWriteResults in a goroutine and
// signals the WaitGroup on completion. Errors are returned via errCh so the
// caller can decide whether to abort or continue.
func SearchAndWriteResultsAsync(
	wg *sync.WaitGroup,
	errCh chan<- error,
	repo *api_results.GHRepo,
	appConfig *config.AppConfig,
	authToken string,
	tokens []string) {
	defer wg.Done()
	if err := SearchAndWriteResults(repo, appConfig, authToken, tokens); err != nil {
		errCh <- err
	}
}

func SearchAndWriteResults(
	repo *api_results.GHRepo,
	appConfig *config.AppConfig,
	authToken string,
	tokens []string) error {
	tokenResultsWriter, err := results.CreateResultsWriter(appConfig.ResultsConfig)
	if err != nil {
		return err
	}

	fmt.Println("Attempting to get repo archive for repo", repo.Name)
	var archiveFile string
	err = utils.Retry(3, 5*time.Second, func() error {
		var rerr error
		archiveFile, rerr = repo.GetRepoArchive(authToken, api_results.Tarball, appConfig.ExtractDir)
		return rerr
	})
	if err != nil {
		return err
	}

	directory, err := utils.ExtractGZIP(archiveFile, appConfig.ExtractDir)
	utils.RemoveFile(archiveFile)
	if err != nil {
		return err
	}
	defer utils.RemoveDir(directory)

	langFiles, err := utils.GetFilesByExtension(directory, appConfig.GetLanguageExts())
	if err != nil {
		return err
	}

	tokenRefs, err := FindTokenRefsInFiles(langFiles, tokens, directory)
	if err != nil {
		return err
	}

	fmt.Println("Saving token search results for repo", repo.Name)
	tokenResults := tokenRefs.ToTokenResults(repo.FullName, repo.Url, repo.DefaultBranch)
	return tokenResultsWriter.WriteTokenResults(tokenResults)
}
