package api_results

import (
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

// archiveClient is used for downloading repository archives. The 5-minute
// timeout prevents goroutines from hanging indefinitely on a stalled download.
var archiveClient = &http.Client{Timeout: 5 * time.Minute}

type GHRepo struct {
	Name              string       `json:"name"`
	FullName          string       `json:"fullName"`
	Description       string       `json:"description"`
	Languages         []GHLanguage `json:"languages"`
	Owner             string       `json:"owner"`
	Url               string       `json:"url"`
	DefaultBranch     string       `json:"defaultBranch"`
	LastModified      time.Time    `json:"lastModified"`
	Dependency        string       `json:"dependency"`
	DependencyVersion string       `json:"dependencyVersion"`
	Directory         string       `json:"directory"`
	APIUrl            string       `json:"apiUrl"`
}

type GHRepoResults struct {
	Repos GHRepos `json:"repos"`
	Count int     `json:"count"`
}

type GHRepoWriteResult struct {
	Repo         string    `dynamodbav:"repo" json:"repo"`
	Sk           string    `dynamodbav:"sk" json:"sk"`
	Dependency   string    `dynamodbav:"dependency" json:"dependency"`
	Version      string    `dynamodbav:"version" json:"version"`
	Url          string    `dynamodbav:"url" json:"url"`
	Directory    string    `dynamodbav:"directory" json:"directory"`
	ScmSite      string    `dynamodbav:"scm_site" json:"scm_site"`
	LastModified time.Time `dynamodbav:"lastModified" json:"lastModified"`
}

type GHRepos []GHRepo
type GHRepoWriteResults []GHRepoWriteResult
type ArchiveFormat string

const (
	// Tarball specifies an archive in gzipped tar format.
	Tarball ArchiveFormat = "tarball"

	// Zipball specifies an archive in zip format.
	Zipball ArchiveFormat = "zipball"
)

func (ghRepoResults GHRepoWriteResults) SaveRepoResultsToFile(fileName string) error {
	return writer.MarshallAndSave(fileName, ghRepoResults)
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

		return "", fmt.Errorf("error getting repository archive %s", err)
	}

	resp, err := archiveClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error getting repository archive %s", err)
	}

	defer resp.Body.Close()

	// Due to default redirect of the http request the response will
	// contain an attachment with a file name in the Content-Disposition
	// header of the respons
	_, mtm, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	if err != nil {
		return "", fmt.Errorf("error getting repository archive %s", err)
	}

	fileName := filepath.Join(directory, mtm["filename"])
	f, err := os.Create(fileName)
	if err != nil {
		return "", fmt.Errorf("error getting repository archive %s", err)
	}

	defer f.Close()

	// Response body is a gzip file
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error getting repository archive %s", err)
	}

	return fileName, nil
}

func (repos *GHRepos) ToWriteResults() GHRepoWriteResults {
	dynamoRepoResults := make(GHRepoWriteResults, 0)
	for _, r := range *repos {
		dynamoRepoResults = append(dynamoRepoResults, GHRepoWriteResult{
			Repo:         r.FullName,
			Sk:           "DEP#" + r.Dependency,
			Dependency:   r.Dependency,
			Version:      r.DependencyVersion,
			Url:          r.Url,
			Directory:    r.Directory,
			ScmSite:      "GitHub",
			LastModified: r.LastModified,
		})
	}

	return dynamoRepoResults
}
