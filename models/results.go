package models

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type FileResult struct {
	Path string `dynamodbav:"path"`
	Link string `dynamodbav:"link"`
}

type FileResults []FileResult

type TokenResult struct {
	Token    string      `dynamodbav:"component" json:"token"`
	RepoName string      `dynamodbav:"repo" json:"repo"`
	Files    FileResults `dynamodbav:"files" json:"files"`
}

type TokenResults []TokenResult

func NewTokenResult(token, repoName string, results FileResults) TokenResult {
	return TokenResult{
		Token:    token,
		RepoName: repoName,
		Files:    results,
	}
}

func (tokenResult TokenResult) GetKey() map[string]types.AttributeValue {
	repo, err := attributevalue.Marshal(tokenResult.RepoName)
	if err != nil {
		panic(err)
	}

	component, err := attributevalue.Marshal(tokenResult.Token)
	if err != nil {
		panic(err)
	}

	return map[string]types.AttributeValue{"repo": repo, "component": component}
}
