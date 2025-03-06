package main

import (
	"aws-shop-backend/packages/middleware"
	"context"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fileName := request.QueryStringParameters["name"]
	if fileName == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"message": "Missing 'name' query parameter"}`,
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
		}, err
	}

	s3Client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3Client)
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
		Key:    aws.String("uploaded/" + fileName),
	}, func(options *s3.PresignOptions) {
		options.Expires = time.Duration(3 * time.Minute)
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"message": "Failed to generate presigned URL"}`,
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       req.URL,
	}, nil
}

func main() {
	h := middleware.Chain(handler, middleware.AddCorsHeaders())
	lambda.Start(h)
}
