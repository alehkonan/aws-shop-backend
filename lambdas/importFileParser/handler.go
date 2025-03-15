package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func handler(ctx context.Context, s3Event events.S3Event) error {
	for _, record := range s3Event.Records {
		bucketName := record.S3.Bucket.Name
		objKey := record.S3.Object.Key

		obj, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objKey),
		})
		if err != nil {
			return fmt.Errorf("failed to get object, %v", err)
		}

		if err := parseObject(ctx, obj); err != nil {
			_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(objKey),
			})
			return fmt.Errorf("failed to parse object, %v", err)
		}

		if err := moveObject(ctx, bucketName, objKey); err != nil {
			return fmt.Errorf("failed to move file: %v", err)
		}
	}

	return nil
}
