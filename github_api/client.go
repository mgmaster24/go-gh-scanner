package github_api

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// GHClient struct - contains the context and a pointer to a github.Client
// that will be used to execute GitHub API methods
type GHClient struct {
	Client *github.Client
	Ctx    context.Context
}

// Creates a new GitHub Client
func NewClient(token string) *GHClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return &GHClient{Client: github.NewClient(tc), Ctx: ctx}
}
