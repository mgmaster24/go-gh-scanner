package results

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mgmaster24/go-gh-scanner/models"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
	"github.com/mgmaster24/go-gh-scanner/utils"
	"github.com/mgmaster24/go-gh-scanner/writer"
)

// FileResultWriter - used for saving files to a directory
type FileResultsWriter struct {
	// Root folder that results will be written to
	Destination string
}

// Creates a new FileResultsWriter pointer
func NewFileResultsWriter(destination string) *FileResultsWriter {
	return &FileResultsWriter{
		Destination: destination,
	}
}

// FileResultWriter - WriteTokenResults
//
// Writes the token results slice to a file in JSON format
func (fileWriter *FileResultsWriter) WriteTokenResults(results models.TokenResults) error {
	if len(results) <= 0 {
		return nil
	}

	dir := filepath.Join(fileWriter.Destination, "components", results[0].Repo)
	if err := utils.CreateDir(dir); err != nil {
		return err
	}

	for i, result := range results {
		if err := writer.MarshallAndSave(fmt.Sprintf("%s/results_%v.json", dir, i), result); err != nil {
			return err
		}
	}

	return nil
}

// FileResultWriter - WriteRepoResults
//
// Groups results by dependency and writes one JSON file per dependency into
// the Destination directory. Results with no dependency are written to repos.json.
func (fileWriter *FileResultsWriter) WriteRepoResults(results api_results.GHRepoWriteResults) error {
	if len(results) == 0 {
		return nil
	}

	dir := filepath.Join(fileWriter.Destination, "repos")
	if err := utils.CreateDir(dir); err != nil {
		return err
	}

	byDep := make(map[string]api_results.GHRepoWriteResults)
	for _, r := range results {
		key := r.Dependency
		if key == "" {
			key = "repos"
		}
		byDep[key] = append(byDep[key], r)
	}

	for dep, depResults := range byDep {
		fileName := filepath.Join(dir, depToFileName(dep))
		if err := depResults.SaveRepoResultsToFile(fileName); err != nil {
			return err
		}
	}

	return nil
}

// depToFileName converts a dependency name like @m2s2/ng-lib to m2s2-ng-lib.json
func depToFileName(dep string) string {
	name := strings.ReplaceAll(dep, "@", "")
	name = strings.ReplaceAll(name, "/", "-")
	return name + ".json"
}

// Update the file writer destination
func (writer *FileResultsWriter) UpdateDestination(destination string) {
	writer.Destination = destination
}
