package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type stackOutputs struct {
	ScannerInfraStack struct {
		TableName      string `json:"TableName"`
		ApiUrl         string `json:"ApiUrl"`
		ScannerRoleArn string `json:"ScannerRoleArn"`
		InfraRoleArn   string `json:"InfraRoleArn"`
	} `json:"ScannerInfraStack"`
}

type cdkJSON struct {
	Context map[string]json.RawMessage `json:"context"`
}

type appConfig struct {
	ExtractDir         string          `json:"extractDir"`
	PerPage            int             `json:"perPage"`
	PackageFile        string          `json:"packageFile"`
	Owner              string          `json:"owner,omitempty"`
	Dependencies       []string        `json:"dependencies"`
	Languages          []language      `json:"languages"`
	ReposToIgnore      []string        `json:"reposToIgnore"`
	ComponentDiscovery discoveryConfig `json:"componentDiscovery"`
	Results            writerConfig    `json:"resultsWriterConfig"`
}

type language struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
}

type discoveryConfig struct {
	Owner string   `json:"owner"`
	Repo  string   `json:"repo,omitempty"`
	Repos []string `json:"repos,omitempty"`
	Paths []string `json:"paths,omitempty"`
}

type writerConfig struct {
	Destination     string `json:"destination"`
	DestinationType string `json:"destinationType"`
	UseBatch        bool   `json:"useBatch"`
}

func main() {
	outputs, err := readOutputs("cdk-outputs.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading cdk-outputs.json:", err)
		fmt.Fprintln(os.Stderr, "run 'cdk deploy --outputs-file cdk-outputs.json' first")
		os.Exit(1)
	}

	ctx, err := readContext("cdk.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading cdk.json:", err)
		os.Exit(1)
	}

	var scanOwner, discoveryOwner, discoveryRepo, githubRepo, awsRegion string
	var discoveryRepos, discoveryPaths, dependencies []string
	json.Unmarshal(ctx["scanOwner"], &scanOwner)
	json.Unmarshal(ctx["discoveryOwner"], &discoveryOwner)
	json.Unmarshal(ctx["discoveryRepo"], &discoveryRepo)
	json.Unmarshal(ctx["discoveryRepos"], &discoveryRepos)
	json.Unmarshal(ctx["discoveryPaths"], &discoveryPaths)
	json.Unmarshal(ctx["dependencies"], &dependencies)
	json.Unmarshal(ctx["githubRepo"], &githubRepo)
	json.Unmarshal(ctx["awsRegion"], &awsRegion)

	stack := outputs.ScannerInfraStack

	cfg := appConfig{
		ExtractDir:   "temp",
		PerPage:      100,
		PackageFile:  "package.json",
		Owner:        scanOwner,
		Dependencies: dependencies,
		Languages: []language{
			{Name: "TypeScript", Extension: ".ts"},
			{Name: "TypeScript React", Extension: ".tsx"},
			{Name: "JavaScript", Extension: ".js"},
			{Name: "JavaScript React", Extension: ".jsx"},
			{Name: "HTML", Extension: ".html"},
			{Name: "Vue", Extension: ".vue"},
		},
		ReposToIgnore: []string{},
		ComponentDiscovery: discoveryConfig{
			Owner: discoveryOwner,
			Repo:  discoveryRepo,
			Repos: discoveryRepos,
			Paths: discoveryPaths,
		},
		Results: writerConfig{
			Destination:     stack.TableName,
			DestinationType: "table",
			UseBatch:        true,
		},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error marshaling config:", err)
		os.Exit(1)
	}

	outPath := "../config/dynamo-config.json"
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "error writing config:", err)
		os.Exit(1)
	}
	fmt.Println("config/dynamo-config.json written")

	if stack.ApiUrl != "" {
		fmt.Printf("API endpoint: %s\n", stack.ApiUrl)
	}

	setGitHubSecret(githubRepo, "AWS_ROLE_ARN", stack.ScannerRoleArn)
	setGitHubSecret(githubRepo, "AWS_INFRA_ROLE_ARN", stack.InfraRoleArn)
	setGitHubSecret(githubRepo, "AWS_REGION", awsRegion)
}

func setGitHubSecret(repo, name, value string) {
	cmd := exec.Command("gh", "secret", "set", name, "--repo", repo, "--body", value)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "could not set %s via gh CLI: %v\n", name, err)
		fmt.Fprintf(os.Stderr, "set it manually: gh secret set %s --repo %s --body '%s'\n", name, repo, value)
	} else {
		fmt.Printf("GitHub secret %s set on %s\n", name, repo)
	}
}

func readOutputs(path string) (*stackOutputs, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var o stackOutputs
	return &o, json.Unmarshal(data, &o)
}

func readContext(path string) (map[string]json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c cdkJSON
	return c.Context, json.Unmarshal(data, &c)
}
