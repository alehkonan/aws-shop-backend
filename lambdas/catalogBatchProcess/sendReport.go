package main

import (
	"aws-shop-backend/packages/products"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// Sends report about created products to the sns topic
func sendReport(
	ctx context.Context,
	createdProducts *[]products.CreateProductDto,
) error {
	if len(*createdProducts) == 0 {
		return nil
	}

	var report strings.Builder

	report.WriteString("The following products were created:\n\n")
	report.WriteString(fmt.Sprintf(
		"| %-20s | %-30s | %-10s | %-10s |\n",
		"Title",
		"Description",
		"Price",
		"Count",
	))
	report.WriteString(fmt.Sprintf(
		"| %-20s | %-30s | %-10s | %-10s |\n",
		strings.Repeat("-", 20),
		strings.Repeat("-", 30),
		strings.Repeat("-", 10),
		strings.Repeat("-", 10),
	))

	for _, product := range *createdProducts {
		row := fmt.Sprintf(
			"| %-20s | %-30s | %-10.2f | %-10d |\n",
			product.Title,
			product.Description,
			product.Price,
			product.Count,
		)
		report.WriteString(row)
	}

	_, err := snsClient.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(productTopicArn),
		Subject:  aws.String("Report about created products"),
		Message:  aws.String(report.String()),
	})
	if err != nil {
		return err
	}

	return nil
}
