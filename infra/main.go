package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	tableName := app.Node().TryGetContext(jsii.String("tableName")).(string)
	githubRepo := app.Node().TryGetContext(jsii.String("githubRepo")).(string)

	// Auth context — leave empty strings if IDP is not yet configured.
	jwksUri, _ := app.Node().TryGetContext(jsii.String("jwksUri")).(string)
	jwtIssuer, _ := app.Node().TryGetContext(jsii.String("jwtIssuer")).(string)
	jwtAudience, _ := app.Node().TryGetContext(jsii.String("jwtAudience")).(string)

	NewScannerStack(app, "ScannerInfraStack", &ScannerStackProps{
		StackProps: awscdk.StackProps{
			Description: jsii.String("DynamoDB table, API Gateway, and IAM roles for go-gh-scanner"),
		},
		TableName:   tableName,
		GitHubRepo:  githubRepo,
		JwksUri:     jwksUri,
		JwtIssuer:   jwtIssuer,
		JwtAudience: jwtAudience,
	})

	app.Synth(nil)
}
