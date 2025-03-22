package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

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
