package models

import "fmt"

type FileMatch struct {
	File       string
	LineNumber int
	Token      string
}

type FileMatches []FileMatch

// Map of token values to token matches found in a file.
type TokenMatchesMap map[string]FileMatches

// Converts a slice of FileMath to a slice of FileResults
//
// The purpose is to combine repository information to make direct links
// to the location of the token instance in GitHub
func (matches FileMatches) ToFileResults(repoUrl string, branch string) FileResults {
	results := make(FileResults, len(matches))
	for i, m := range matches {
		results[i] = FileResult{
			Path: m.File,
			Link: fmt.Sprintf("%s/blob/%s%s#L%v", repoUrl, branch, m.File, m.LineNumber),
		}
	}

	return results
}

// Constructs a slice of TokenResult from a TokenMatchesMap
//
// Combine repository information and FileMatch information to prepare repository
// specific token search data for writing to its destination
func (matchMap TokenMatchesMap) ToTokenResults(repoName, url, defaultBranch string) TokenResults {
	results := make(TokenResults, 0)
	for k, v := range matchMap {
		results = append(results, NewTokenResult(k, repoName, v.ToFileResults(url, defaultBranch)))
	}

	return results
}
