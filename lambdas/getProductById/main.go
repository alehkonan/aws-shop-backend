package main

import (
	"aws-shop-backend/packages/middleware"
	"aws-shop-backend/packages/products"
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	productId, ok := event.PathParameters["productId"]
	if !ok {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"message": "Product id was not found"}`,
		}, nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"message": "Failed to load AWS config"}`,
		}, nil
	}

	repo := products.Repository(dynamodb.NewFromConfig(cfg))

	data, err := repo.GetProductById(ctx, productId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"message": "Failed to get product from database"}`,
		}, nil
	}

	json, err := json.Marshal(data)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"message": "Failed to parse database response"}`,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(json),
	}, nil
}

func main() {
	h := middleware.Chain(handler, middleware.AddCorsHeaders())
	lambda.Start(h)
}
