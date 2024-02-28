package results

import (
	"fmt"

	"github.com/mgmaster24/go-gh-scanner/aws_sdk"
	"github.com/mgmaster24/go-gh-scanner/config"
	"github.com/mgmaster24/go-gh-scanner/models"
)

// Interface for defining token result saving
type ResultsWriter interface {
	Write(results models.TokenResults) error
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
		return aws_sdk.NewDynamoDBResultsWriter(writerConfig.Destination, writerConfig.UseBatchProcessing), nil
	}

	return nil, fmt.Errorf("no ResultsWriter was created based on the provided configuration. Config %v", writerConfig)
}
