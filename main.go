package main

import (
	"fmt"

	"github.com/mgmaster24/go-gh-scanner/aws_sdk"
	"github.com/mgmaster24/go-gh-scanner/cli"
	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/github_api"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
	"github.com/mgmaster24/go-gh-scanner/search"
	"github.com/mgmaster24/go-gh-scanner/tokens"
	"github.com/mgmaster24/go-gh-scanner/utils"
	"github.com/mgmaster24/go-gh-scanner/writer/results"
)

func main() {
	flags := cli.InitFlags()
	appConfig, err := config.Read(flags.AppConfig)
	panicIfError(err)

	// Get secrets service client
	secretsService := aws_sdk.NewSecretsManagerClient()
	// Get the auth token value from the secret service
	authToken, err := secretsService.GetSecretString(appConfig.AuthTokenKey)
	panicIfError(err)

	// Create the github api client
	client := github_api.NewClient(authToken)
	// Run scan to get results
	scanResults, err := client.ScanPackageDeps(appConfig)
	panicIfError(err)

	fmt.Println("Getting repository data for repos found during dependency scan.")
	ghRepoResults, err := scanResults.ToRepoData(appConfig.Owner, appConfig.TeamsToIgnore, client.GetRepoData)
	panicIfError(err)

	fmt.Println("Saving repository results")
	writer, err := results.CreateResultsWriter(appConfig.RepoResultsConfig)
	panicIfError(err)

	panicIfError(writer.WriteRepoResults(ghRepoResults.Repos.ToWriteResults()))
	panicIfError(utils.CreateDir(appConfig.ExtractDir))

	// get tokens - This is just and example of how to read tokens of different types for search
	tokenRetriever, err := tokens.CreateTokenReader(tokens.NGTokenReaderType)
	panicIfError(err)
	panicIfError(tokenRetriever.Fetch(flags.TokensConfig))

	writer.UpdateDestination(appConfig.TokenResultsConfig.Destination)
	for _, repo := range ghRepoResults.Repos {
		fmt.Println("Attempting to get repo archive for repo", repo.Name)
		archiveFile, err := repo.GetRepoArchive(authToken, api_results.Tarball, appConfig.ExtractDir)
		panicIfError(fmt.Errorf("error getting repository archive %s", err))

		// extract gzip
		directory, err := utils.ExtractGZIP(archiveFile, appConfig.ExtractDir)
		panicIfError(err)
		// get files for language extensions
		langFiles, err := utils.GetFilesByExtension(directory, appConfig.GetLanguageExts())
		panicIfError(err)
		// search language specific files
		tokenRefs, err := search.FindTokenRefsInFiles(langFiles, tokenRetriever.ToTokens(), directory)
		panicIfError(err)

		// write tokens search results
		fmt.Println("Saving token search results for repo", repo.Name)
		tokenResults := tokenRefs.ToTokenResults(repo.FullName, repo.Url, repo.DefaultBranch)
		panicIfError(writer.WriteTokenResults(tokenResults))

		// Remove extracted directory
		panicIfError(utils.RemoveDir(directory))
		// Remove the archive file
		panicIfError(utils.RemoveFile(archiveFile))
	}

	fmt.Println("Scan operation complete.  Exiting...")
}

func panicIfError(e error) {
	if e != nil {
		panic(e)
	}
}
