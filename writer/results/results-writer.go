package results

import (
	"fmt"

	"github.com/mgmaster24/go-gh-scanner/aws_sdk/dynamodb_client"
	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/models"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
)

// Interface for defining result saving features
type ResultsWriter interface {
	WriteTokenResults(results models.TokenResults) error
	WriteRepoResults(results api_results.GHRepoWriteResults) error
	UpdateDestination(destination string)
}

// Create a writer for the destination type provided in the writer config.
//
// Implemented Destination Types:
//
//   - File
//   - Table (Currently writes to DynamoDB)
func CreateResultsWriter(writerConfig config.WriterConfig) (ResultsWriter, error) {
	switch writerConfig.DestinationType {
	case config.FileDestination:
		return NewFileResultsWriter(writerConfig.Destination), nil
	case config.TableDesitnation:
		return dynamodb_client.NewDynamoDBClient(writerConfig.Destination, writerConfig.UseBatchProcessing), nil
	}

	return nil, fmt.Errorf("no ResultsWriter was created based on the provided configuration. Config %v", writerConfig)
}
