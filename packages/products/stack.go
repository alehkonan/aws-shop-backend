package products

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssnssubscriptions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	productsTable := awsdynamodb.Table_FromTableName(stack, jsii.String("ProductsTable"), jsii.String("products"))
	stocksTable := awsdynamodb.Table_FromTableName(stack, jsii.String("StocksTable"), jsii.String("stocks"))

	catalogItemsQueue := awssqs.NewQueue(
		stack,
		jsii.String("CatalogItemsQueue"),
		&awssqs.QueueProps{
			QueueName: jsii.String("catalogItemsQueue"),
		},
	)

	createProductTopic := awssns.NewTopic(
		stack, jsii.String("CreateProductTopic"),
		&awssns.TopicProps{
			TopicName: jsii.String("createProductTopic"),
		},
	)

	createProductTopic.AddSubscription(
		awssnssubscriptions.NewEmailSubscription(jsii.String("o.konan@softteco.com"), nil),
	)

	// Subscription to high price products
	createProductTopic.AddSubscription(
		awssnssubscriptions.NewEmailSubscription(
			jsii.String("o.konan+highprice@softteco.com"),
			&awssnssubscriptions.EmailSubscriptionProps{
				FilterPolicy: &map[string]awssns.SubscriptionFilter{
					"price": awssns.SubscriptionFilter_NumericFilter(
						&awssns.NumericConditions{
							GreaterThan: aws.Float64(100),
						},
					),
				},
			},
		),
	)

	getProductListFunction := awslambda.NewFunction(stack, jsii.String("getProductListFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Code:    awslambda.Code_FromAsset(jsii.String("lambdas/getProductList"), nil),
		Handler: jsii.String("bootstrap"),
		Environment: &map[string]*string{
			"PRODUCTS_TABLE": productsTable.TableName(),
			"STOCKS_TABLE":   stocksTable.TableName(),
		},
	})

	getProductByIdFunction := awslambda.NewFunction(stack, jsii.String("GetProductByIdFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Code:    awslambda.Code_FromAsset(jsii.String("lambdas/getProductById"), nil),
		Handler: jsii.String("bootstrap"),
		Environment: &map[string]*string{
			"PRODUCTS_TABLE": productsTable.TableName(),
			"STOCKS_TABLE":   stocksTable.TableName(),
		},
	})

	createProductFunction := awslambda.NewFunction(stack, jsii.String("CreateProductFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Code:    awslambda.Code_FromAsset(jsii.String("lambdas/createProduct"), nil),
		Handler: jsii.String("bootstrap"),
		Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
		Environment: &map[string]*string{
			"PRODUCTS_TABLE": productsTable.TableName(),
			"STOCKS_TABLE":   stocksTable.TableName(),
		},
	})

	productsTable.GrantReadData(getProductListFunction)
	productsTable.GrantReadData(getProductByIdFunction)
	productsTable.GrantReadWriteData(createProductFunction)
	stocksTable.GrantReadData(getProductListFunction)
	stocksTable.GrantReadData(getProductByIdFunction)
	stocksTable.GrantReadWriteData(createProductFunction)

	productApi := awsapigateway.NewRestApi(stack, jsii.String("ProductApi"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String("Product Api"),
		DeployOptions: &awsapigateway.StageOptions{
			StageName: jsii.String("dev"),
		},
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowOrigins: awsapigateway.Cors_ALL_ORIGINS(),
		},
	})

	products := productApi.Root().AddResource(jsii.String("products"), nil)
	products.AddMethod(
		jsii.String("GET"),
		awsapigateway.NewLambdaIntegration(getProductListFunction, nil),
		nil,
	)
	products.AddMethod(
		jsii.String("POST"),
		awsapigateway.NewLambdaIntegration(createProductFunction, nil),
		nil,
	)

	product := products.AddResource(jsii.String("{productId}"), nil)
	product.AddMethod(
		jsii.String("GET"),
		awsapigateway.NewLambdaIntegration(getProductByIdFunction, nil),
		nil,
	)

	catalogBatchProcessFunction := awslambda.NewFunction(
		stack,
		jsii.String("CatalogBatchProcessFunction"),
		&awslambda.FunctionProps{
			Runtime: awslambda.Runtime_PROVIDED_AL2023(),
			Code:    awslambda.Code_FromAsset(jsii.String("lambdas/catalogBatchProcess"), nil),
			Handler: jsii.String("bootstrap"),
			Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
			Environment: &map[string]*string{
				"PRODUCTS_TABLE":    productsTable.TableName(),
				"STOCKS_TABLE":      stocksTable.TableName(),
				"PRODUCT_TOPIC_ARN": createProductTopic.TopicArn(),
			},
		},
	)

	catalogBatchProcessFunction.AddEventSource(
		awslambdaeventsources.NewSqsEventSource(
			catalogItemsQueue, &awslambdaeventsources.SqsEventSourceProps{
				BatchSize: jsii.Number(5),
			},
		),
	)

	productsTable.GrantReadWriteData(catalogBatchProcessFunction)
	stocksTable.GrantReadWriteData(catalogBatchProcessFunction)
	catalogItemsQueue.GrantConsumeMessages(catalogBatchProcessFunction)
	createProductTopic.GrantPublish(catalogBatchProcessFunction)

	awscdk.NewCfnOutput(stack, jsii.String("ProductApiUrl"), &awscdk.CfnOutputProps{
		Value:       productApi.Url(),
		Description: jsii.String("Product API Gateway URL"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("CatalogItemsQueueArn"), &awscdk.CfnOutputProps{
		Value:      catalogItemsQueue.QueueArn(),
		ExportName: jsii.String("CatalogItemsQueueArn"),
	})

	return stack
}
