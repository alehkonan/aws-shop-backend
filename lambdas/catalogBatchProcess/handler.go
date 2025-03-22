package main

import (
	"aws-shop-backend/packages/helpers"
	"aws-shop-backend/packages/products"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
)

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	var createdProducts []products.CreateProductDto

	for _, record := range sqsEvent.Records {
		message := strings.Split(record.Body, ",")
		product := products.CreateProductDto{
			Title:       message[0],
			Description: message[1],
			Price:       helpers.ConvertPrice(message[2]),
			Count:       helpers.ConvertCount(message[3]),
		}

		_, err := productRepo.CreateProduct(ctx, product)
		if err != nil {
			log.Printf("can not create product, %v", err)
			continue
		}

		if product.Price > 100 {
			_, err := snsClient.Publish(ctx, &sns.PublishInput{
				TopicArn: aws.String(productTopicArn),
				Subject:  aws.String("Product with high price is created"),
				Message:  aws.String(fmt.Sprintf("Product details: %v", product)),
				MessageAttributes: map[string]types.MessageAttributeValue{
					"price": {
						DataType:    aws.String("Number"),
						StringValue: aws.String(message[2]),
					},
				},
			})
			if err != nil {
				log.Printf("can not publish product, %v", err)
				continue
			}
		}

		createdProducts = append(createdProducts, product)
	}

	if err := sendReport(ctx, &createdProducts); err != nil {
		return err
	}

	return nil
}
