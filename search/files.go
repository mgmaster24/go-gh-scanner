package search

import (
	"bufio"
	"io"
	"strings"

	"github.com/mgmaster24/go-gh-scanner/models"
	"github.com/mgmaster24/go-gh-scanner/utils"
)

// Open a file and search each line for instances of the provided token
//
// The returned FileMathches will contain relative paths to the searched file,
// the line numbers of the associated occurrence and the token that was used
// to search on.
func findTokenOccurences(filePath string, token string, dir string) (*models.FileMatches, error) {
	file, err := utils.OpenFile(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	matches := make(models.FileMatches, 0)
	reader := bufio.NewReader(file)
	lineNum := 1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		if strings.Contains(line, token) {
			matches = append(matches, models.FileMatch{
				File:       filePath[len(dir):],
				LineNumber: lineNum,
				Token:      token,
			})
		}

		lineNum++
	}

	return &matches, nil
}

// Run through the provided tokens and search each provided file for occurrences of the token
//
// Returns a TokenMatchesMap that maps to token that was search to a slice of FileMatch instances
func FindTokenRefsInFiles(files []string, tokens []string, dir string) (models.TokenMatchesMap, error) {
	tokenMatches := make(models.TokenMatchesMap)
	for _, token := range tokens {
		fileMatches := make(models.FileMatches, 0)
		for _, file := range files {
			fms, err := findTokenOccurences(file, token, dir)
			if err != nil {
				return nil, err
			}

			if len(*fms) > 0 {
				fileMatches = append(fileMatches, *fms...)
			}
		}

		if len(fileMatches) > 0 {
			tokenMatches[token] = fileMatches
		}
	}

	return tokenMatches, nil
}
