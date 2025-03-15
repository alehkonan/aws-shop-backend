package main

import (
	"aws-shop-backend/packages/helpers"
	"aws-shop-backend/packages/products"
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
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
			return err
		}

		createdProducts = append(createdProducts, product)
	}

	if err := sendReport(ctx, &createdProducts); err != nil {
		return err
	}

	return nil
}
