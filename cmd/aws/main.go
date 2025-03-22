package main

import (
	"aws-shop-backend/packages/imports"
	"aws-shop-backend/packages/products"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	productsStack := products.NewStack(app, "ProductsStack", &awscdk.StackProps{
		Env: env(),
	})

	importsStack := imports.NewStack(app, "ImportsStack", &awscdk.StackProps{
		Env: env(),
	})

	importsStack.AddDependency(productsStack, jsii.String("Products stack must be created first"))

	app.Synth(nil)
}
