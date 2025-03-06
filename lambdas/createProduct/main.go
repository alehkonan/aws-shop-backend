package main

import (
	"aws-shop-backend/packages/middleware"
	"aws-shop-backend/packages/products"
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var newProduct products.CreateProductDto
	if err := json.Unmarshal([]byte(event.Body), &newProduct); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"message": "Invalid request body"}`,
		}, nil
	}

	validate := validator.New()
	if err := validate.Struct(newProduct); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: fmt.Sprintf(`{"message": "Validation error", "errors": %s}`, validationErrors.Error()),
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

	data, err := repo.CreateProduct(ctx, newProduct)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"message": "Failed to create new product"}`,
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
		StatusCode: 201,
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
