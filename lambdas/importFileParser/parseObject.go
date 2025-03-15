package main

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Parses s3 object with csv reader
func parseObject(ctx context.Context, obj *s3.GetObjectOutput) error {
	defer obj.Body.Close()

	reader := csv.NewReader(obj.Body)

	// read the header not to include it to the output
	_, err := reader.Read()
	if err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Errors will be logged in CloudWatch
			log.Printf("%s", err.Error())
			continue
		}

		_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:    aws.String(catalogQueueUrl),
			MessageBody: aws.String(strings.Join(record, ",")),
		})
		if err != nil {
			log.Printf("error sending message to SQS: %v", err)
			continue
		}
	}

	return nil
}
