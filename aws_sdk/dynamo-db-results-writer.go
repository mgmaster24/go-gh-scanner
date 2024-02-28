package aws_sdk

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mgmaster24/go-gh-scanner/models"
)

// DynamoDBResultWriter struct
//
// Contains the destination and dynamodb client for interacting with DynamoDB
type DynamoDBResultsWriter struct {
	TableName string
	Client    *dynamodb.Client
	UseBatch  bool
}

// Attempts to create DynamoDB writer by loading the default AWS config
// and creating a new dynamodb client from the config.  Also provides the
// table name of the table items should be written to
func NewDynamoDBResultsWriter(tableName string, useBatch bool) *DynamoDBResultsWriter {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		panic(err)
	}

	ddbClient := dynamodb.NewFromConfig(sdkConfig)
	ddbWriter := &DynamoDBResultsWriter{Client: ddbClient}
	ddbWriter.TableName = tableName
	ddbWriter.UseBatch = useBatch
	return ddbWriter
}

// Batch write the Token Results to the table name provided in the
// DynamoDBResultsWriter
func (ddbw *DynamoDBResultsWriter) Write(results models.TokenResults) error {
	if ddbw.UseBatch {
		return ddbw.writeBatch(results)
	} else {
		return ddbw.writePutItems(results)
	}
}

func (ddbw *DynamoDBResultsWriter) writeBatch(results models.TokenResults) error {
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
				return fmt.Errorf("couldn't marshall TokenResult %v for batch writing. Here's why: %v", result.Token, err)
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

func (ddbw *DynamoDBResultsWriter) writePutItems(results models.TokenResults) error {
	for _, result := range results {
		item, err := attributevalue.MarshalMap(result)
		if err != nil {
			return fmt.Errorf("couldn't marshall TokenResult %v for batch writing. Here's why: %v", result.Token, err)
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
				return fmt.Errorf("couldn't put the item %v in %v table. Here's why: %v", result.Token, ddbw.TableName, err)
			}

			break
		}
	}

	return nil
}
