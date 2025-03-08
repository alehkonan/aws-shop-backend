package main

import (
	"aws-shop-backend/packages/middleware"
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	uploadPrefix = "uploaded/"
)

var (
	presignClient *s3.PresignClient
	bucket        string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	presignClient = s3.NewPresignClient(s3Client)

	bucket = os.Getenv("BUCKET_NAME")
	if bucket == "" {
		log.Fatalf("bucket name is not defined")
	}
}

// Generates url for file uploading
func getPresignedUrl(ctx context.Context, fileName string) (string, error) {
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(uploadPrefix + fileName),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(1 * time.Minute)
	})

	return req.URL, err
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	url, err := getPresignedUrl(ctx, fileName)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"message": "Failed to generate url for file uploading"}`,
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       url,
	}, nil
}

func main() {
	handler := middleware.Chain(handleRequest, middleware.AddCorsHeaders())
	lambda.Start(handler)
}
