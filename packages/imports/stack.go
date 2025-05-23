package imports

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3notifications"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	uploadBucket := awss3.Bucket_FromBucketName(stack, jsii.String("UploadBucket"), jsii.String("aws-shop-uploads"))

	catalogItemsQueue := awssqs.Queue_FromQueueAttributes(stack, jsii.String("CatalogItemsQueue"), &awssqs.QueueAttributes{
		QueueArn: awscdk.Fn_ImportValue(jsii.String("CatalogItemsQueueArn")),
	})

	importProductsFileFunction := awslambda.NewFunction(stack, jsii.String("ImportProductsFileFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Code:    awslambda.Code_FromAsset(jsii.String("lambdas/importProductsFile"), nil),
		Handler: jsii.String("bootstrap"),
		Environment: &map[string]*string{
			"BUCKET_NAME": uploadBucket.BucketName(),
		},
	})
	uploadBucket.GrantReadWrite(importProductsFileFunction, nil)

	importFileParserFunction := awslambda.NewFunction(stack, jsii.String("ImportFileParserFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Code:    awslambda.Code_FromAsset(jsii.String("lambdas/importFileParser"), nil),
		Handler: jsii.String("bootstrap"),
		Environment: &map[string]*string{
			"BUCKET_NAME":       uploadBucket.BucketName(),
			"CATALOG_QUEUE_URL": catalogItemsQueue.QueueUrl(),
		},
	})
	catalogItemsQueue.GrantSendMessages(importFileParserFunction)
	uploadBucket.GrantReadWrite(importFileParserFunction, nil)

	uploadBucket.AddEventNotification(awss3.EventType_OBJECT_CREATED,
		awss3notifications.NewLambdaDestination(importFileParserFunction),
		&awss3.NotificationKeyFilter{
			Prefix: jsii.String("uploaded/"),
		},
	)

	basicAuthorizerFunction := awslambda.Function_FromFunctionAttributes(
		stack, jsii.String("BasicAuthorizerFunction"), &awslambda.FunctionAttributes{
			FunctionArn:     awscdk.Fn_ImportValue(jsii.String("BasicAuthorizerFunctionArn")),
			SameEnvironment: jsii.Bool(true),
		},
	)

	basicAuthorizer := awsapigateway.NewTokenAuthorizer(
		stack, jsii.String("TokenAuthorizer"),
		&awsapigateway.TokenAuthorizerProps{
			Handler:        basicAuthorizerFunction,
			AuthorizerName: jsii.String("TokenAuthorizer"),
			IdentitySource: jsii.String("method.request.header.Authorization"),
		},
	)

	importApi := awsapigateway.NewRestApi(stack, jsii.String("ImportApi"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String("Import Api"),
		DeployOptions: &awsapigateway.StageOptions{
			StageName: jsii.String("dev"),
		},
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowOrigins: awsapigateway.Cors_ALL_ORIGINS(),
		},
	})

	importApi.AddGatewayResponse(jsii.String("Unauthorized"), &awsapigateway.GatewayResponseOptions{
		Type:       awsapigateway.ResponseType_UNAUTHORIZED(),
		StatusCode: jsii.String("401"),
		ResponseHeaders: &map[string]*string{
			"Access-Control-Allow-Origin": jsii.String("'*'"),
		},
	})

	importApi.AddGatewayResponse(jsii.String("Forbidden"), &awsapigateway.GatewayResponseOptions{
		Type:       awsapigateway.ResponseType_ACCESS_DENIED(),
		StatusCode: jsii.String("403"),
		ResponseHeaders: &map[string]*string{
			"Access-Control-Allow-Origin": jsii.String("'*'"),
		},
	})

	importResource := importApi.Root().AddResource(jsii.String("import"), nil)
	importResource.AddMethod(jsii.String("GET"),
		awsapigateway.NewLambdaIntegration(importProductsFileFunction, nil),
		&awsapigateway.MethodOptions{
			Authorizer: basicAuthorizer,
			RequestParameters: &map[string]*bool{
				"method.request.querystring.name": jsii.Bool(true),
			},
		},
	)

	return stack
}
