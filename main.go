package main

import (
	"fmt"
	"sync"

	"github.com/mgmaster24/go-gh-scanner/aws_sdk"
	"github.com/mgmaster24/go-gh-scanner/cli"
	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/github_api"
	"github.com/mgmaster24/go-gh-scanner/search"
	"github.com/mgmaster24/go-gh-scanner/tokens"
	"github.com/mgmaster24/go-gh-scanner/utils"
	"github.com/mgmaster24/go-gh-scanner/writer/results"
)

func main() {
	flags := cli.InitFlags()
	appConfig, err := config.Read(flags.AppConfig)
	utils.PanicIfError(err)

	// Get secrets service client
	secretsService := aws_sdk.NewSecretsManagerClient()
	// Get the auth token value from the secret service
	authToken, err := secretsService.GetSecretString(appConfig.AuthTokenKey)
	utils.PanicIfError(err)

	// Create the github api client
	client := github_api.NewClient(authToken)
	// Run scan to get results
	scanResults, err := client.ScanPackageDeps(appConfig)
	scanResults = scanResults.RemoveDuplicates()
	utils.PanicIfError(err)

	fmt.Println("Getting repository data for repos found during dependency scan.")
	ghRepoResults, err := scanResults.ToRepoData(appConfig.Owner, appConfig.TeamsToIgnore, client.GetRepoData)
	utils.PanicIfError(err)

	fmt.Println("Saving repository results")
	writer, err := results.CreateResultsWriter(appConfig.RepoResultsConfig)
	utils.PanicIfError(err)

	utils.PanicIfError(writer.WriteRepoResults(ghRepoResults.Repos.ToWriteResults()))
	utils.PanicIfError(utils.CreateDir(appConfig.ExtractDir))

	// get tokens - This is just and example of how to read tokens of different types for search
	tokenRetriever, err := tokens.CreateTokenReader(tokens.NGTokenReaderType)
	utils.PanicIfError(err)
	utils.PanicIfError(tokenRetriever.Fetch(flags.TokensConfig))
	tokens := tokenRetriever.ToTokens()

	var waitGroup sync.WaitGroup
	for _, repo := range ghRepoResults.Repos {
		waitGroup.Add(1)
		go search.SearchAndWriteResultsAsync(&waitGroup, &repo, appConfig, authToken, tokens)
	}

	waitGroup.Wait()

	fmt.Println("Scan operation complete.  Exiting...")
}
