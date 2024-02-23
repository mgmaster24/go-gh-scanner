package search

import (
	"bufio"
	"io"
	"path/filepath"
	"strings"

	"github.com/mgmaster24/go-gh-scanner/utils"
)

type FileMatches struct {
	File       string
	LineNumber int
	Line       string
	Token      string
}

func FindTokenOccurences(filePath string, token string) (*[]FileMatches, error) {
	file, err := utils.OpenFile(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	matches := make([]FileMatches, 0)
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
			matches = append(matches, FileMatches{
				File:       filePath,
				LineNumber: lineNum,
				Line:       line,
				Token:      token,
			})
		}

		lineNum++
	}

	return &matches, nil
}

func FindTokenRefsInFiles(files []string, tokens []string) (map[string][]FileMatches, error) {
	tokenMatches := make(map[string][]FileMatches)
	for _, file := range files {
		fileMatches := make([]FileMatches, 0)
		for _, token := range tokens {
			fms, err := FindTokenOccurences(file, token)
			if err != nil {
				return nil, err
			}

			if len(*fms) > 0 {
				fileMatches = append(fileMatches, *fms...)
			}
		}

		if len(fileMatches) > 0 {
			tokenMatches[filepath.Base(file)] = fileMatches
		}
	}

	return tokenMatches, nil
}
