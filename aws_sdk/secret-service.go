package aws_sdk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type SecretsManClient struct {
	Client *secretsmanager.Client
}

func NewSecretsManagerClient() *SecretsManClient {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		panic(err)
	}

	smClient := secretsmanager.NewFromConfig(sdkConfig)
	return &SecretsManClient{
		Client: smClient,
	}
}

func (smClient *SecretsManClient) GetSecretString(secretKey string) (string, error) {
	output, err := smClient.Client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: &secretKey,
	})

	if err != nil {
		return "", err
	}

	var secret map[string]string
	err = json.Unmarshal([]byte(*output.SecretString), &secret)
	if err != nil {
		return "", err
	}
	return secret[secretKey], nil
}
