package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type dbClient struct {
	ddb       *dynamodb.Client
	tableName string
}

func newDBClient(ddb *dynamodb.Client, tableName string) *dbClient {
	return &dbClient{ddb: ddb, tableName: tableName}
}

type DepRecord struct {
	Repo         string    `dynamodbav:"repo" json:"repo"`
	Dependency   string    `dynamodbav:"dependency" json:"dependency"`
	Version      string    `dynamodbav:"version" json:"version"`
	Url          string    `dynamodbav:"url" json:"url"`
	Directory    string    `dynamodbav:"directory" json:"directory"`
	ScmSite      string    `dynamodbav:"scm_site" json:"scm_site"`
	LastModified time.Time `dynamodbav:"lastModified" json:"lastModified"`
}

type CompRecord struct {
	Repo      string       `dynamodbav:"repo" json:"repo"`
	Component string       `dynamodbav:"component" json:"component"`
	Files     []FileResult `dynamodbav:"files" json:"files"`
}

type FileResult struct {
	Path string `dynamodbav:"path" json:"path"`
	Link string `dynamodbav:"link" json:"link"`
}

// queryDependencyIndex returns all dep records for a given npm dependency.
func (c *dbClient) queryDependencyIndex(ctx context.Context, dependency string) ([]DepRecord, error) {
	resp, err := c.ddb.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(c.tableName),
		IndexName:              aws.String("dependency-index"),
		KeyConditionExpression: aws.String("dependency = :dep"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":dep": &types.AttributeValueMemberS{Value: dependency},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dependency-index query: %w", err)
	}
	var records []DepRecord
	if err := attributevalue.UnmarshalListOfMaps(resp.Items, &records); err != nil {
		return nil, fmt.Errorf("unmarshal dep records: %w", err)
	}
	return records, nil
}

// queryComponentIndex returns all component records for a given component name.
func (c *dbClient) queryComponentIndex(ctx context.Context, component string) ([]CompRecord, error) {
	resp, err := c.ddb.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(c.tableName),
		IndexName:              aws.String("component-index"),
		KeyConditionExpression: aws.String("component = :comp"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":comp": &types.AttributeValueMemberS{Value: component},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("component-index query: %w", err)
	}
	var records []CompRecord
	if err := attributevalue.UnmarshalListOfMaps(resp.Items, &records); err != nil {
		return nil, fmt.Errorf("unmarshal component records: %w", err)
	}
	return records, nil
}

// queryRepo returns all dep and component records for the given repo (PK query).
func (c *dbClient) queryRepo(ctx context.Context, repoName string) ([]DepRecord, []CompRecord, error) {
	resp, err := c.ddb.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(c.tableName),
		KeyConditionExpression: aws.String("#r = :repo"),
		ExpressionAttributeNames: map[string]string{
			"#r": "repo",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":repo": &types.AttributeValueMemberS{Value: repoName},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("repo query: %w", err)
	}

	var deps []DepRecord
	var comps []CompRecord
	for _, item := range resp.Items {
		sk, _ := item["sk"].(*types.AttributeValueMemberS)
		if sk == nil {
			continue
		}
		switch {
		case strings.HasPrefix(sk.Value, "DEP#"):
			var r DepRecord
			if err := attributevalue.UnmarshalMap(item, &r); err == nil {
				deps = append(deps, r)
			}
		case strings.HasPrefix(sk.Value, "COMP#"):
			var r CompRecord
			if err := attributevalue.UnmarshalMap(item, &r); err == nil {
				comps = append(comps, r)
			}
		}
	}
	return deps, comps, nil
}

// scanRepos returns a page of unique repo names. Pagination is cursor-based;
// pass an empty cursor for the first page. Note: deduplication is per-page only.
func (c *dbClient) scanRepos(ctx context.Context, cursor string) (repos []string, next string, err error) {
	input := &dynamodb.ScanInput{
		TableName:            aws.String(c.tableName),
		FilterExpression:     aws.String("begins_with(sk, :prefix)"),
		ProjectionExpression: aws.String("#r"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":prefix": &types.AttributeValueMemberS{Value: "DEP#"},
		},
		ExpressionAttributeNames: map[string]string{
			"#r": "repo",
		},
	}
	if cursor != "" {
		if startKey, decErr := decodeCursor(cursor); decErr == nil {
			input.ExclusiveStartKey = startKey
		}
	}

	resp, err := c.ddb.Scan(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("repo scan: %w", err)
	}

	seen := make(map[string]bool, len(resp.Items))
	for _, item := range resp.Items {
		if r, ok := item["repo"].(*types.AttributeValueMemberS); ok && !seen[r.Value] {
			seen[r.Value] = true
			repos = append(repos, r.Value)
		}
	}

	if len(resp.LastEvaluatedKey) > 0 {
		next = encodeCursor(resp.LastEvaluatedKey)
	}
	return repos, next, nil
}

type cursorKey struct {
	Repo string `json:"repo"`
	Sk   string `json:"sk"`
}

func encodeCursor(key map[string]types.AttributeValue) string {
	k := cursorKey{}
	if v, ok := key["repo"].(*types.AttributeValueMemberS); ok {
		k.Repo = v.Value
	}
	if v, ok := key["sk"].(*types.AttributeValueMemberS); ok {
		k.Sk = v.Value
	}
	data, _ := json.Marshal(k)
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeCursor(cursor string) (map[string]types.AttributeValue, error) {
	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}
	var k cursorKey
	if err := json.Unmarshal(data, &k); err != nil {
		return nil, err
	}
	return map[string]types.AttributeValue{
		"repo": &types.AttributeValueMemberS{Value: k.Repo},
		"sk":   &types.AttributeValueMemberS{Value: k.Sk},
	}, nil
}
