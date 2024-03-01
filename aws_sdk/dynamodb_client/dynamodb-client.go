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

// DynamoDBResultWriter struct
//
// Contains the destination and dynamodb client for interacting with DynamoDB
type DynamoDBClient struct {
	TableName string
	Client    *dynamodb.Client
	UseBatch  bool
}

// Attempts to create DynamoDB writer by loading the default AWS config
// and creating a new dynamodb client from the config.  Also provides the
// table name of the table items should be written to
func NewDynamoDBClient(tableName string, useBatch bool) *DynamoDBClient {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		panic(err)
	}

	ddbClient := dynamodb.NewFromConfig(sdkConfig)
	ddbWriter := &DynamoDBClient{Client: ddbClient}
	ddbWriter.TableName = tableName
	ddbWriter.UseBatch = useBatch
	return ddbWriter
}

func writeBatch[T any](ddbw *DynamoDBClient, results []T) error {
	written := 0
	batchSize := 10
	start := 0
	end := start + batchSize
	lenResults := len(results)
	for start < lenResults {
		var writeRequests []types.WriteRequest
		if end > lenResults {
			end = lenResults
		}

		for _, result := range results[start:end] {
			item, err := attributevalue.MarshalMap(result)
			if err != nil {
				return fmt.Errorf("couldn't marshall valeu %v for batch writing. Here's why: %v", result, err)
			} else {
				writeRequests = append(writeRequests, types.WriteRequest{PutRequest: &types.PutRequest{Item: item}})
			}
		}

		for {
			_, err := ddbw.Client.BatchWriteItem(
				context.TODO(),
				&dynamodb.BatchWriteItemInput{
					RequestItems: map[string][]types.WriteRequest{ddbw.TableName: writeRequests}})
			if err != nil {
				if _, ok := err.(*types.ProvisionedThroughputExceededException); ok {
					fmt.Println("exceeded the maximum write capacity units.  Retrying after 3 seconds...")
					time.Sleep(3 * time.Second)
					continue
				}
				return fmt.Errorf("couldn't add a batch of TokenResults to %v. Here's why: %v", ddbw.TableName, err)
			} else {
				written += len(writeRequests)
				break
			}
		}

		start = end
		end += batchSize
	}

	return nil
}

func writeItem[T any](ddbw *DynamoDBClient, results []T) error {
	for _, result := range results {
		item, err := attributevalue.MarshalMap(result)
		if err != nil {
			return fmt.Errorf("couldn't marshall result %v for batch writing. Here's why: %v", result, err)
		}

		for {
			_, err = ddbw.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
				TableName: aws.String(ddbw.TableName), Item: item,
			})
			if err != nil {
				fmt.Println(err)
				if _, ok := err.(*types.RequestLimitExceeded); ok {
					fmt.Println("exceeded the maximum write capacity units.  Retrying after 3 seconds...")
					time.Sleep(3 * time.Second)
					continue
				}
				return fmt.Errorf("couldn't put the item %v in %v table. Here's why: %v", result, ddbw.TableName, err)
			}

			break
		}
	}

	return nil
}
