package dynamodb_client

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBClient wraps a DynamoDB client with a target table name.
type DynamoDBClient struct {
	TableName string
	Client    *dynamodb.Client
	UseBatch  bool
}

// NewDynamoDBClient loads the default AWS config and returns a DynamoDBClient
// targeting the given table.
func NewDynamoDBClient(tableName string, useBatch bool) *DynamoDBClient {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		panic(err)
	}

	return &DynamoDBClient{
		Client:    dynamodb.NewFromConfig(sdkConfig),
		TableName: tableName,
		UseBatch:  useBatch,
	}
}

// deleteByKeyPrefix queries every item whose partition key equals pkValue and
// whose sort key begins with skPrefix, then deletes them in batches of 25.
// Pagination is handled automatically so large result sets are fully cleared.
func deleteByKeyPrefix(ddbw *DynamoDBClient, pkAttr, pkValue, skAttr, skPrefix string) error {
	ctx := context.TODO()
	var lastKey map[string]types.AttributeValue

	for {
		resp, err := ddbw.Client.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(ddbw.TableName),
			KeyConditionExpression: aws.String("#pk = :pk AND begins_with(#sk, :prefix)"),
			ExpressionAttributeNames: map[string]string{
				"#pk": pkAttr,
				"#sk": skAttr,
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":     &types.AttributeValueMemberS{Value: pkValue},
				":prefix": &types.AttributeValueMemberS{Value: skPrefix},
			},
			ProjectionExpression: aws.String("#pk, #sk"),
			ExclusiveStartKey:    lastKey,
		})
		if err != nil {
			return fmt.Errorf("error querying stale items (pk=%s value=%s sk_prefix=%s): %w", pkAttr, pkValue, skPrefix, err)
		}

		if len(resp.Items) > 0 {
			if err := deleteItems(ddbw, resp.Items, pkAttr, skAttr); err != nil {
				return err
			}
		}

		lastKey = resp.LastEvaluatedKey
		if len(lastKey) == 0 {
			break
		}
	}

	return nil
}

// deleteItems deletes the provided items in batches of 25, retrying
// any unprocessed items until all are removed.
func deleteItems(ddbw *DynamoDBClient, items []map[string]types.AttributeValue, pkAttr, skAttr string) error {
	const batchSize = 25
	ctx := context.TODO()

	for start := 0; start < len(items); start += batchSize {
		end := min(start+batchSize, len(items))

		pending := make([]types.WriteRequest, 0, end-start)
		for _, item := range items[start:end] {
			pending = append(pending, types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: map[string]types.AttributeValue{
						pkAttr: item[pkAttr],
						skAttr: item[skAttr],
					},
				},
			})
		}

		for len(pending) > 0 {
			resp, err := ddbw.Client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]types.WriteRequest{ddbw.TableName: pending},
			})
			if err != nil {
				if _, ok := err.(*types.ProvisionedThroughputExceededException); ok {
					time.Sleep(3 * time.Second)
					continue
				}
				return fmt.Errorf("error deleting batch from %s: %w", ddbw.TableName, err)
			}
			// Retry any items DynamoDB could not process in this round.
			pending = resp.UnprocessedItems[ddbw.TableName]
		}
	}

	return nil
}

// writeBatch writes results in batches of 25 and retries any unprocessed items.
func writeBatch[T any](ddbw *DynamoDBClient, results []T) error {
	const batchSize = 25
	ctx := context.TODO()

	for start := 0; start < len(results); start += batchSize {
		end := min(start+batchSize, len(results))

		pending := make([]types.WriteRequest, 0, end-start)
		for _, result := range results[start:end] {
			item, err := attributevalue.MarshalMap(result)
			if err != nil {
				return fmt.Errorf("couldn't marshal value for batch write: %w", err)
			}
			pending = append(pending, types.WriteRequest{PutRequest: &types.PutRequest{Item: item}})
		}

		for len(pending) > 0 {
			resp, err := ddbw.Client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]types.WriteRequest{ddbw.TableName: pending},
			})
			if err != nil {
				if _, ok := err.(*types.ProvisionedThroughputExceededException); ok {
					fmt.Println("write capacity exceeded — retrying in 3 seconds...")
					time.Sleep(3 * time.Second)
					continue
				}
				return fmt.Errorf("couldn't batch write to %s: %w", ddbw.TableName, err)
			}
			pending = resp.UnprocessedItems[ddbw.TableName]
		}
	}

	return nil
}

func writeItem[T any](ddbw *DynamoDBClient, results []T) error {
	ctx := context.TODO()
	for _, result := range results {
		item, err := attributevalue.MarshalMap(result)
		if err != nil {
			return fmt.Errorf("couldn't marshal result for PutItem: %w", err)
		}

		for {
			_, err = ddbw.Client.PutItem(ctx, &dynamodb.PutItemInput{
				TableName: aws.String(ddbw.TableName),
				Item:      item,
			})
			if err != nil {
				if _, ok := err.(*types.RequestLimitExceeded); ok {
					fmt.Println("request limit exceeded — retrying in 3 seconds...")
					time.Sleep(3 * time.Second)
					continue
				}
				return fmt.Errorf("couldn't put item in %s: %w", ddbw.TableName, err)
			}
			break
		}
	}

	return nil
}
