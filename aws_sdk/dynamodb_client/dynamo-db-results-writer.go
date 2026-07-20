package dynamodb_client

import (
	"github.com/mgmaster24/go-gh-scanner/models"
	"github.com/mgmaster24/go-gh-scanner/models/api_results"
)

// WriteRepoResults deletes all existing DEP# records for each repo present in
// results, then writes the fresh set. Grouping by repo (the PK) ensures we
// only remove the repo's own stale dependency entries, never touching
// component records that share the same partition.
func (ddbc *DynamoDBClient) WriteRepoResults(results api_results.GHRepoWriteResults) error {
	seen := make(map[string]bool)
	for _, r := range results {
		if !seen[r.Repo] {
			seen[r.Repo] = true
			if err := deleteByKeyPrefix(ddbc, "repo", r.Repo, "sk", "DEP#"); err != nil {
				return err
			}
		}
	}
	return write(ddbc, results)
}

// WriteTokenResults deletes all existing COMP# records for each repo present in
// results, then writes the fresh set. This ensures component usage entries
// from repos that dropped a component never persist across runs.
func (ddbc *DynamoDBClient) WriteTokenResults(results models.TokenResults) error {
	seen := make(map[string]bool)
	for _, r := range results {
		if !seen[r.Repo] {
			seen[r.Repo] = true
			if err := deleteByKeyPrefix(ddbc, "repo", r.Repo, "sk", "COMP#"); err != nil {
				return err
			}
		}
	}
	return write(ddbc, results)
}

func write[T any](ddbc *DynamoDBClient, results []T) error {
	if ddbc.UseBatch {
		return writeBatch(ddbc, results)
	}
	return writeItem(ddbc, results)
}

func (writer *DynamoDBClient) UpdateDestination(destination string) {
	writer.TableName = destination
}
