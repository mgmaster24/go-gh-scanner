package models

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
