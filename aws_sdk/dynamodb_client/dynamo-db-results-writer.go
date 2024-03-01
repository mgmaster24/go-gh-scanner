package dynamodb_client

import (
	"github.com/mgmaster24/go-gh-scanner/models"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
)

// Batch write the Token Results to the table name provided in the
// DynamoDBResultsWriter
func (ddbc *DynamoDBClient) WriteTokenResults(results models.TokenResults) error {
	if ddbc.UseBatch {
		return writeBatch(ddbc, results)
	} else {
		return writeItem(ddbc, results)
	}
}

func (ddbc *DynamoDBClient) WriteRepoResults(results api_results.GHRepoDynamoResults) error {
	return writeBatch(ddbc, results)
}

func (writer *DynamoDBClient) UpdateDestination(destination string) {
	writer.TableName = destination
}
