package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type server struct {
	db   *dbClient
	auth *jwksCache
}

func main() {
	sdkCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load AWS config:", err)
		os.Exit(1)
	}

	srv := &server{
		db: newDBClient(
			dynamodb.NewFromConfig(sdkCfg),
			mustEnv("TABLE_NAME"),
		),
		auth: newJWKSCache(
			os.Getenv("JWKS_URI"),
			os.Getenv("JWT_ISSUER"),
			os.Getenv("JWT_AUDIENCE"),
		),
	}

	lambda.Start(srv.handle)
}

func (s *server) handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if _, err := s.auth.validate(req.Headers["authorization"]); err != nil {
		return jsonResp(http.StatusUnauthorized, map[string]string{"error": "unauthorized"}), nil
	}

	path := req.RawPath
	q := req.QueryStringParameters

	switch {
	case path == "/repos" && q["dependency"] != "":
		return s.reposByDependency(ctx, q["dependency"])
	case path == "/repos" && q["component"] != "":
		return s.reposByComponent(ctx, q["component"])
	case path == "/repos":
		return s.listRepos(ctx, q["next"])
	case path == "/repo" && q["name"] != "":
		return s.repoDetail(ctx, q["name"])
	default:
		return jsonResp(http.StatusNotFound, map[string]string{"error": "not found"}), nil
	}
}

func (s *server) reposByDependency(ctx context.Context, dep string) (events.APIGatewayV2HTTPResponse, error) {
	records, err := s.db.queryDependencyIndex(ctx, dep)
	if err != nil {
		return jsonResp(http.StatusInternalServerError, errBody(err)), nil
	}
	if records == nil {
		records = []DepRecord{}
	}
	return jsonResp(http.StatusOK, map[string]any{"repos": records}), nil
}

func (s *server) reposByComponent(ctx context.Context, comp string) (events.APIGatewayV2HTTPResponse, error) {
	records, err := s.db.queryComponentIndex(ctx, comp)
	if err != nil {
		return jsonResp(http.StatusInternalServerError, errBody(err)), nil
	}
	if records == nil {
		records = []CompRecord{}
	}
	return jsonResp(http.StatusOK, map[string]any{"repos": records}), nil
}

func (s *server) repoDetail(ctx context.Context, name string) (events.APIGatewayV2HTTPResponse, error) {
	deps, comps, err := s.db.queryRepo(ctx, name)
	if err != nil {
		return jsonResp(http.StatusInternalServerError, errBody(err)), nil
	}
	if deps == nil {
		deps = []DepRecord{}
	}
	if comps == nil {
		comps = []CompRecord{}
	}
	return jsonResp(http.StatusOK, map[string]any{
		"repo":         name,
		"dependencies": deps,
		"components":   comps,
	}), nil
}

func (s *server) listRepos(ctx context.Context, cursor string) (events.APIGatewayV2HTTPResponse, error) {
	repos, next, err := s.db.scanRepos(ctx, cursor)
	if err != nil {
		return jsonResp(http.StatusInternalServerError, errBody(err)), nil
	}
	if repos == nil {
		repos = []string{}
	}
	body := map[string]any{"repos": repos}
	if next != "" {
		body["next"] = next
	}
	return jsonResp(http.StatusOK, body), nil
}

func jsonResp(status int, body any) events.APIGatewayV2HTTPResponse {
	data, _ := json.Marshal(body)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type":                  "application/json",
			"Access-Control-Allow-Origin":   "*",
			"Access-Control-Allow-Headers":  "Authorization, Content-Type",
		},
		Body: string(data),
	}
}

func errBody(err error) map[string]string {
	return map[string]string{"error": err.Error()}
}

func mustEnv(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	fmt.Fprintf(os.Stderr, "required environment variable %s is not set\n", key)
	os.Exit(1)
	return ""
}
