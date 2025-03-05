package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Moves file from /uploaded to /parsed folder
func moveFile(ctx context.Context, s3Client *s3.Client, bucket, sourceKey string) error {
	if !strings.HasPrefix(sourceKey, "uploaded/") {
		return fmt.Errorf("source key does not have 'uploaded/' prefix: %s", sourceKey)
	}

	targetKey := "parsed/" + sourceKey[len("uploaded/"):]

	_, err := s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &bucket,
		CopySource: aws.String(fmt.Sprintf("%s/%s", bucket, sourceKey)),
		Key:        &targetKey,
	})
	if err != nil {
		return fmt.Errorf("copy operation failed: %v", err)
	}

	_, err = s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &targetKey,
	})
	if err != nil {
		return fmt.Errorf("failed to verify copied object: %v", err)
	}

	_, err = s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &sourceKey,
	})
	if err != nil {
		return fmt.Errorf("delete operation failed: %v", err)
	}

	waiter := s3.NewObjectNotExistsWaiter(s3Client)
	err = waiter.Wait(ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &sourceKey,
	}, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to confirm deletion: %v", err)
	}

	return nil
}
