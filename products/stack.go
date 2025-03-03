package products

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	productsTable := awsdynamodb.Table_FromTableName(stack, jsii.String("ProductsTable"), jsii.String("products"))
	stocksTable := awsdynamodb.Table_FromTableName(stack, jsii.String("StocksTable"), jsii.String("stocks"))

	getProductListFunction := awslambda.NewFunction(stack, jsii.String("getProductListFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_GO_1_X(),
		Code:    awslambda.Code_FromAsset(jsii.String("handlers/getProductList"), nil),
		Handler: jsii.String("bootstrap"),
		Environment: &map[string]*string{
			"PRODUCTS_TABLE": productsTable.TableName(),
			"STOCKS_TABLE":   stocksTable.TableName(),
		},
	})

	getProductByIdFunction := awslambda.NewFunction(stack, jsii.String("GetProductByIdFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_GO_1_X(),
		Code:    awslambda.Code_FromAsset(jsii.String("handlers/getProductById"), nil),
		Handler: jsii.String("bootstrap"),
		Environment: &map[string]*string{
			"PRODUCTS_TABLE": productsTable.TableName(),
			"STOCKS_TABLE":   stocksTable.TableName(),
		},
	})

	createProductFunction := awslambda.NewFunction(stack, jsii.String("CreateProductFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_GO_1_X(),
		Code:    awslambda.Code_FromAsset(jsii.String("handlers/createProduct"), nil),
		Handler: jsii.String("bootstrap"),
		Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
		Environment: &map[string]*string{
			"PRODUCTS_TABLE": productsTable.TableName(),
			"STOCKS_TABLE":   stocksTable.TableName(),
		},
	})

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

	awscdk.NewCfnOutput(stack, jsii.String("ProductApiUrl"), &awscdk.CfnOutputProps{
		Value:       productApi.Url(),
		Description: jsii.String("Product API Gateway URL"),
	})

	return stack
}
