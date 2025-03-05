package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func handler(ctx context.Context, s3Event events.S3Event) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	for _, record := range s3Event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key
		result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &key,
		})
		if err != nil {
			return fmt.Errorf("failed to get object %s/%s: %v", bucket, key, err)
		}

		defer result.Body.Close()

		reader := csv.NewReader(result.Body)
		header, err := reader.Read()
		if err != nil {
			return fmt.Errorf("failed to read CSV header: %v", err)
		}
		fmt.Printf("CSV Headers: %v\n", header)

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error reading CSV record: %v", err)
			}

			fmt.Printf("Processing record: %v\n", record)
		}

		if err := moveFile(ctx, s3Client, bucket, key); err != nil {
			return fmt.Errorf("failed to move file: %v", err)
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
