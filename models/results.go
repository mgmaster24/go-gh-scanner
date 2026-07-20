package models

type FileResult struct {
	Path string `dynamodbav:"path"`
	Link string `dynamodbav:"link"`
}

type FileResults []FileResult

type TokenResult struct {
	Repo      string      `dynamodbav:"repo" json:"repo"`
	Sk        string      `dynamodbav:"sk" json:"sk"`
	Component string      `dynamodbav:"component" json:"component"`
	Files     FileResults `dynamodbav:"files" json:"files"`
}

type TokenResults []TokenResult

func NewTokenResult(token, repoName string, results FileResults) TokenResult {
	return TokenResult{
		Repo:      repoName,
		Sk:        "COMP#" + token,
		Component: token,
		Files:     results,
	}
}
