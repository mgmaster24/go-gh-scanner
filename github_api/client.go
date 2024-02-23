package github_api

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GHClient struct {
	Client *github.Client
	Ctx    context.Context
}

func NewClient(token string) *GHClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return &GHClient{Client: github.NewClient(tc), Ctx: ctx}
}
