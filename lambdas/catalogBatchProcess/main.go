package main

import (
	"aws-shop-backend/packages/products"
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SnsClient interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

var (
	snsClient       SnsClient
	productRepo     *products.ProductRepository
	productTopicArn string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	productRepo = products.Repository(dynamodb.NewFromConfig(cfg))
	snsClient = sns.NewFromConfig(cfg)
	productTopicArn = os.Getenv("PRODUCT_TOPIC_ARN")
}

func main() {
	lambda.Start(handler)
}
