package aws_sdk

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func ListBuckets() {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return
	}
	fmt.Println(sdkConfig)
	s3Client := s3.NewFromConfig(sdkConfig)
	results, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	for _, b := range results.Buckets {
		fmt.Println(*b.Name)
	}
}
