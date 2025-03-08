package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	uploadPrefix = "uploaded/"
	parsePrefix  = "parsed/"
)

type S3Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

var (
	s3Client S3Client
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	s3Client = s3.NewFromConfig(cfg)
}

// Parses s3 object with csv reader
func parseCsv(obj *s3.GetObjectOutput) error {
	defer obj.Body.Close()

	reader := csv.NewReader(obj.Body)

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

	return nil
}

// Moves s3 object from /uploaded to /parsed
func moveObject(ctx context.Context, bucketName string, objKey string) error {
	targetKey := parsePrefix + objKey[len(uploadPrefix):]

	_, err := s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(bucketName),
		CopySource: aws.String(fmt.Sprintf("%s/%s", bucketName, objKey)),
		Key:        aws.String(targetKey),
	})
	if err != nil {
		return err
	}

	_, err = s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objKey),
	})
	if err != nil {
		return err
	}

	return nil
}

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

		if err := parseCsv(obj); err != nil {
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

func main() {
	lambda.Start(handler)
}
