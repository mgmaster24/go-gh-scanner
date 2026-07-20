package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ScannerStackProps struct {
	awscdk.StackProps
	TableName   string
	GitHubRepo  string
	JwksUri     string
	JwtIssuer   string
	JwtAudience string
}

func NewScannerStack(scope constructs.Construct, id string, props *ScannerStackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	table := awsdynamodb.NewTable(stack, jsii.String("ScannerTable"), &awsdynamodb.TableProps{
		TableName: jsii.String(props.TableName),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("repo"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("sk"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
		RemovalPolicy: awscdk.RemovalPolicy_RETAIN,
	})

	// GSI: answers "which repos use @m2s2/ng-lib?"
	table.AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
		IndexName: jsii.String("dependency-index"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("dependency"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("repo"),
			Type: awsdynamodb.AttributeType_STRING,
		},
	})

	// GSI: answers "which repos use m2s2-button?"
	table.AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
		IndexName: jsii.String("component-index"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("component"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("repo"),
			Type: awsdynamodb.AttributeType_STRING,
		},
	})

	// Import the existing GitHub Actions OIDC provider — only one can exist per AWS account.
	oidcProvider := awsiam.OpenIdConnectProvider_FromOpenIdConnectProviderArn(
		stack,
		jsii.String("GitHubOIDC"),
		jsii.String("arn:aws:iam::"+*stack.Account()+":oidc-provider/token.actions.githubusercontent.com"),
	)

	principal := awsiam.NewOpenIdConnectPrincipal(oidcProvider, &map[string]any{
		"StringEquals": map[string]any{
			"token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
		},
		"StringLike": map[string]any{
			"token.actions.githubusercontent.com:sub": "repo:" + props.GitHubRepo + ":*",
		},
	})

	// Scanner role — DynamoDB read/write only, used by the weekly scan workflow.
	scannerRole := awsiam.NewRole(stack, jsii.String("ScannerRole"), &awsiam.RoleProps{
		AssumedBy:   principal,
		Description: jsii.String("Assumed by the scanner workflow to read/write DynamoDB results"),
	})

	scannerRole.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: &[]*string{
			jsii.String("dynamodb:Query"),
			jsii.String("dynamodb:BatchWriteItem"),
			jsii.String("dynamodb:PutItem"),
			jsii.String("dynamodb:DeleteItem"),
		},
		Resources: &[]*string{
			table.TableArn(),
			jsii.String(*table.TableArn() + "/index/*"),
		},
	}))

	// Infra role — CloudFormation + IAM + DynamoDB, used by the deploy-infra workflow.
	infraRole := awsiam.NewRole(stack, jsii.String("InfraRole"), &awsiam.RoleProps{
		AssumedBy:   principal,
		Description: jsii.String("Assumed by the infra deploy workflow to manage CDK stacks"),
	})

	infraRole.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: &[]*string{
			jsii.String("cloudformation:*"),
			jsii.String("dynamodb:CreateTable"),
			jsii.String("dynamodb:DeleteTable"),
			jsii.String("dynamodb:DescribeTable"),
			jsii.String("dynamodb:UpdateTable"),
			jsii.String("dynamodb:UpdateContinuousBackups"),
			jsii.String("iam:CreateRole"),
			jsii.String("iam:DeleteRole"),
			jsii.String("iam:GetRole"),
			jsii.String("iam:AttachRolePolicy"),
			jsii.String("iam:DetachRolePolicy"),
			jsii.String("iam:PutRolePolicy"),
			jsii.String("iam:DeleteRolePolicy"),
			jsii.String("iam:PassRole"),
			jsii.String("iam:CreateOpenIDConnectProvider"),
			jsii.String("iam:GetOpenIDConnectProvider"),
			jsii.String("iam:DeleteOpenIDConnectProvider"),
			jsii.String("iam:TagOpenIDConnectProvider"),
			jsii.String("lambda:*"),
			jsii.String("apigateway:*"),
			jsii.String("logs:*"),
			jsii.String("ssm:GetParameter"),
			jsii.String("s3:GetObject"),
			jsii.String("s3:PutObject"),
			jsii.String("s3:ListBucket"),
		},
		Resources: &[]*string{jsii.String("*")},
	}))

	// API Lambda function — read-only DynamoDB access, JWT-protected via JWKS.
	apiFunc := awslambda.NewFunction(stack, jsii.String("ApiFunction"), &awslambda.FunctionProps{
		FunctionName: jsii.String(props.TableName + "-api"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Architecture: awslambda.Architecture_ARM_64(),
		Handler:      jsii.String("bootstrap"),
		Code:         awslambda.Code_FromAsset(jsii.String("../api/dist"), nil),
		Environment: &map[string]*string{
			"TABLE_NAME":   jsii.String(props.TableName),
			"JWKS_URI":     jsii.String(props.JwksUri),
			"JWT_ISSUER":   jsii.String(props.JwtIssuer),
			"JWT_AUDIENCE": jsii.String(props.JwtAudience),
		},
		Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
	})

	table.GrantReadData(apiFunc)

	// HTTP API Gateway — lightweight, lower cost than REST API.
	httpApi := awsapigatewayv2.NewHttpApi(stack, jsii.String("ScannerApi"), &awsapigatewayv2.HttpApiProps{
		ApiName: jsii.String(props.TableName + "-api"),
		CorsPreflight: &awsapigatewayv2.CorsPreflightOptions{
			AllowOrigins: &[]*string{jsii.String("*")},
			AllowHeaders: &[]*string{
				jsii.String("Authorization"),
				jsii.String("Content-Type"),
			},
			AllowMethods: &[]awsapigatewayv2.CorsHttpMethod{
				awsapigatewayv2.CorsHttpMethod_GET,
			},
		},
	})

	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("ApiIntegration"),
		apiFunc,
		nil,
	)

	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path:        jsii.String("/{proxy+}"),
		Methods:     &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_GET},
		Integration: integration,
	})

	awscdk.NewCfnOutput(stack, jsii.String("TableName"), &awscdk.CfnOutputProps{
		Value:       table.TableName(),
		Description: jsii.String("Set as resultsWriterConfig.destination in your scanner config"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ApiUrl"), &awscdk.CfnOutputProps{
		Value:       httpApi.ApiEndpoint(),
		Description: jsii.String("API Gateway endpoint for querying scan results"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ScannerRoleArn"), &awscdk.CfnOutputProps{
		Value:       scannerRole.RoleArn(),
		Description: jsii.String("Set as AWS_ROLE_ARN in GitHub Actions secrets"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("InfraRoleArn"), &awscdk.CfnOutputProps{
		Value:       infraRole.RoleArn(),
		Description: jsii.String("Set as AWS_INFRA_ROLE_ARN in GitHub Actions secrets"),
	})

	return stack
}
