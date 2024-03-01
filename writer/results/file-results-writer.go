package results

import (
	"fmt"

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
	// Might have zero results
	if len(results) <= 0 {
		return nil
	}

	// create directory for repo results
	dir := fileWriter.Destination + "/" + results[0].RepoName
	err := utils.CreateDir(dir)
	if err != nil {
		return err
	}

	for i, result := range results {
		err := writer.MarshallAndSave(fmt.Sprintf("%s/results_%v.json", dir, i), result)
		if err != nil {
			return err
		}
	}

	return nil
}

// FileResultWriter - WriteRepoResults
//
// Writes the repo results slice to a file in JSON format
func (fileWriter *FileResultsWriter) WriteRepoResults(results api_results.GHRepoDynamoResults) error {
	return results.SaveRepoResultsToFile(fileWriter.Destination)
}

// Update the file writer destination
func (writer *FileResultsWriter) UpdateDestination(destination string) {
	writer.Destination = destination
}
