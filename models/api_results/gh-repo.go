package api_results

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mgmaster24/go-gh-scanner/utils"
	"github.com/mgmaster24/go-gh-scanner/writer"
)

type GHRepo struct {
	Name               string       `json:"name"`
	FullName           string       `json:"fullName"`
	Description        string       `json:"description"`
	Languages          []GHLanguage `json:"languages"`
	Owner              string       `json:"owner"`
	Url                string       `json:"url"`
	Team               string       `json:"team"`
	DefaultBranch      string       `json:"defaultBranch"`
	LastModified       time.Time    `json:"lastModified"`
	DependencyVersions string       `json:"dependencyVersion"`
	APIUrl             string       `json:"apiUrl"`
	HTMLUrl            string       `json:"htppUrl"`
}

type GHRepos []GHRepo

type GHRepoResults struct {
	Repos GHRepos `json:"repos"`
	Count int     `json:"count"`
}

type ArchiveFormat string

const (
	// Tarball specifies an archive in gzipped tar format.
	Tarball ArchiveFormat = "tarball"

	// Zipball specifies an archive in zip format.
	Zipball ArchiveFormat = "zipball"
)

func (ghRepoResults *GHRepoResults) SaveRepoResultsToFile(fileName string) error {
	return writer.MarshallAndSave(fileName, ghRepoResults)
}

func (ghRepoResults *GHRepoResults) SaveRepoResultsToAWS() error {
	return errors.New("not implemented yet")
}

func (repo *GHRepo) GetRepoArchive(token string, archiveFmt ArchiveFormat, directory string) (string, error) {
	acceptEncoding := "gzip"
	if archiveFmt == Zipball {
		acceptEncoding = "zip"
	}

	req, err := utils.NewHttpRequestNoBody(
		http.MethodGet,
		fmt.Sprintf("%s/%s/%s", repo.APIUrl, archiveFmt, repo.DefaultBranch),
		&map[string]string{
			"Authorization":   "Bearer " + token,
			"Accept":          "application/vnd.github+json",
			"Accept-Encoding": acceptEncoding,
		},
	)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Due to default redirect of the http request the response will
	// contain an attachment with a file name in the Content-Disposition
	// header of the respons
	_, mtm, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	if err != nil {
		return "", err
	}

	fileName := filepath.Join(directory, mtm["filename"])
	f, err := os.Create(fileName)
	if err != nil {
		return "", err
	}

	defer f.Close()

	// Response body is a gzip file
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", err
	}

	return fileName, nil
}
