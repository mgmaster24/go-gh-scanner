package main

import (
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

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
	start := time.Now()

	flags := cli.InitFlags()
	appConfig, err := config.Read(flags.AppConfig)
	if err != nil {
		slog.Error("failed to read config", "error", err)
		os.Exit(1)
	}

	// Ensure the extract directory is cleaned up on exit regardless of outcome.
	if err := utils.CreateDir(appConfig.ExtractDir); err != nil {
		slog.Error("failed to create extract directory", "path", appConfig.ExtractDir, "error", err)
		os.Exit(1)
	}
	defer utils.RemoveDir(appConfig.ExtractDir)

	var authToken utils.Secret
	if raw := os.Getenv("GITHUB_TOKEN"); raw != "" {
		authToken = utils.NewSecret(raw)
	} else {
		if appConfig.AuthTokenKey == "" {
			slog.Error("either GITHUB_TOKEN environment variable or authTokenKey in config is required")
			os.Exit(1)
		}
		secretsService := aws_sdk.NewSecretsManagerClient()
		raw, err := secretsService.GetSecretString(appConfig.AuthTokenKey)
		if err != nil {
			slog.Error("failed to retrieve auth token from Secrets Manager", "key", appConfig.AuthTokenKey, "error", err)
			os.Exit(1)
		}
		authToken = utils.NewSecret(raw)
	}

	client := github_api.NewClient(authToken.Value())

	scope := "all public repos"
	if appConfig.Owner != "" {
		scope = appConfig.Owner
	}
	slog.Info("scanning for dependency usage", "dependencies", appConfig.Dependencies, "scope", scope)
	scanResults, err := client.ScanPackageDeps(appConfig)
	if err != nil {
		slog.Error("dependency scan failed", "error", err)
		os.Exit(1)
	}
	scanResults = scanResults.RemoveDuplicates()
	slog.Info("dependency scan complete", "repos_found", len(scanResults))

	slog.Info("fetching repository metadata")
	ghRepoResults, err := scanResults.ToRepoDataAsync(client.GetRepoData)
	if err != nil {
		slog.Error("failed to fetch repository metadata", "error", err)
		os.Exit(1)
	}

	slog.Info("writing repository results", "count", ghRepoResults.Count)
	repoWriter, err := results.CreateResultsWriter(appConfig.ResultsConfig)
	if err != nil {
		slog.Error("failed to create repo results writer", "error", err)
		os.Exit(1)
	}
	if err := repoWriter.WriteRepoResults(ghRepoResults.Repos.ToWriteResults()); err != nil {
		slog.Error("failed to write repo results", "error", err)
		os.Exit(1)
	}

	var toks []string
	disc := appConfig.ComponentDiscovery
	if disc.Repo != "" {
		slog.Info("auto-discovering components from monorepo", "repo", disc.Repo, "paths", disc.Paths)
		toks, err = client.DiscoverComponentTokensFromMonorepo(
			disc.Owner, disc.Repo, disc.Paths,
			appConfig.ExtractDir, authToken.Value(),
		)
		if err != nil {
			slog.Error("component discovery failed", "error", err)
			os.Exit(1)
		}
		slog.Info("component discovery complete", "tokens_found", len(toks))
	} else if len(disc.Repos) > 0 {
		slog.Info("auto-discovering components from source repos", "repos", disc.Repos)
		toks, err = client.DiscoverComponentTokens(
			disc.Owner, disc.Repos,
			appConfig.ExtractDir, authToken.Value(),
		)
		if err != nil {
			slog.Error("component discovery failed", "error", err)
			os.Exit(1)
		}
		slog.Info("component discovery complete", "tokens_found", len(toks))
	} else {
		tokenRetriever, err := tokens.CreateTokenReader(tokens.JSONTokenReaderType)
		if err != nil {
			slog.Error("failed to create token reader", "error", err)
			os.Exit(1)
		}
		if err := tokenRetriever.Fetch(flags.TokensConfig); err != nil {
			slog.Error("failed to load tokens file", "path", flags.TokensConfig, "error", err)
			os.Exit(1)
		}
		toks = tokenRetriever.ToTokens()
		slog.Info("loaded tokens from file", "count", len(toks), "path", flags.TokensConfig)
	}

	// Limit concurrent archive downloads and searches to avoid memory pressure
	// and GitHub secondary rate limits.
	const workerCount = 10
	sem := make(chan struct{}, workerCount)
	errCh := make(chan error, len(ghRepoResults.Repos))
	total := len(ghRepoResults.Repos)
	var completed atomic.Int32

	var wg sync.WaitGroup
	for _, repo := range ghRepoResults.Repos {
		wg.Add(1)
		sem <- struct{}{}
		go func(r api_results.GHRepo) {
			defer func() {
				n := completed.Add(1)
				slog.Info("repo search progress", "completed", n, "total", total, "repo", r.Name)
				<-sem
			}()
			search.SearchAndWriteResultsAsync(&wg, errCh, &r, appConfig, authToken.Value(), toks)
		}(repo)
	}

	wg.Wait()
	close(errCh)

	var searchErrs int
	for err := range errCh {
		slog.Error("repo search error", "error", err)
		searchErrs++
	}

	slog.Info("scan complete", "elapsed", time.Since(start).Round(time.Second), "search_errors", searchErrs)
}
