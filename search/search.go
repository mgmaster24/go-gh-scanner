package search

import (
	"fmt"
	"sync"

	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
	"github.com/mgmaster24/go-gh-scanner/utils"
	"github.com/mgmaster24/go-gh-scanner/writer/results"
)

func SearchAndWriteResultsAsync(
	waitGroup *sync.WaitGroup,
	repo *api_results.GHRepo,
	appConfig *config.AppConfig,
	authToken string,
	tokens []string) {
	defer waitGroup.Done()
	SearchAndWriteResults(repo, appConfig, authToken, tokens)
}

func SearchAndWriteResults(
	repo *api_results.GHRepo,
	appConfig *config.AppConfig,
	authToken string,
	tokens []string) {
	tokenResultsWriter, err := results.CreateResultsWriter(appConfig.TokenResultsConfig)
	utils.PanicIfError(err)
	fmt.Println("Attempting to get repo archive for repo", repo.Name)
	archiveFile, err := repo.GetRepoArchive(authToken, api_results.Tarball, appConfig.ExtractDir)
	utils.PanicIfError(err)

	// extract gzip
	directory, err := utils.ExtractGZIP(archiveFile, appConfig.ExtractDir)
	utils.PanicIfError(err)
	// get files for language extensions
	langFiles, err := utils.GetFilesByExtension(directory, appConfig.GetLanguageExts())
	utils.PanicIfError(err)
	// search language specific files
	tokenRefs, err := FindTokenRefsInFiles(langFiles, tokens, directory)
	utils.PanicIfError(err)

	// write tokens search results
	fmt.Println("Saving token search results for repo", repo.Name)
	tokenResults := tokenRefs.ToTokenResults(repo.FullName, repo.Url, repo.DefaultBranch)
	utils.PanicIfError(tokenResultsWriter.WriteTokenResults(tokenResults))

	// Remove extracted directory
	utils.PanicIfError(utils.RemoveDir(directory))
	// Remove the archive file
	utils.PanicIfError(utils.RemoveFile(archiveFile))
}
