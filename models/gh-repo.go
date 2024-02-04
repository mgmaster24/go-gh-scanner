package models

type GHRepo struct {
	Name        string `json:"title"`
	Description string `json:"description"`
	ContentsUrl string
	Language    string
}
