package main

import (
	"aws-shop-backend/products"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

var mockProducts = []products.ProductDto{
	{
		Id:          uuid.NewString(),
		Title:       "Apple",
		Description: "Apple from Poland",
		Price:       1.5,
		Count:       199,
	},
	{
		Id:          uuid.NewString(),
		Title:       "Pineapple",
		Description: "Apple from Africa",
		Price:       2.7,
		Count:       120,
	},
	{
		Id:          uuid.NewString(),
		Title:       "Avocado",
		Description: "Green avocado",
		Price:       4.1,
		Count:       49,
	},
	{
		Id:          uuid.NewString(),
		Title:       "Banana",
		Description: "Big banana",
		Price:       0.9,
		Count:       137,
	},
}

func SeedProducts() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Panicf("Unable to load SDK config: %v", err)
		return
	}
	client := dynamodb.NewFromConfig(cfg)

	fmt.Println("Start batch seeding of products...")

	var productWriteRequests []types.WriteRequest
	var stockWriteRequests []types.WriteRequest

	for _, mockProduct := range mockProducts {
		product := products.Product{
			Id:          mockProduct.Id,
			Title:       mockProduct.Title,
			Description: mockProduct.Description,
			Price:       mockProduct.Price,
		}

		stock := products.Stock{
			ProductId: mockProduct.Id,
			Count:     mockProduct.Count,
		}

		productItem, err := attributevalue.MarshalMap(product)
		if err != nil {
			log.Printf("Error marshalling product: %v", err)
			continue
		}

		stockItem, err := attributevalue.MarshalMap(stock)
		if err != nil {
			log.Printf("Error marshalling stock: %v", err)
			continue
		}

		productWriteRequests = append(productWriteRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: productItem,
			},
		})

		stockWriteRequests = append(stockWriteRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: stockItem,
			},
		})
	}

	_, err = client.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			"products": productWriteRequests,
			"stocks":   stockWriteRequests,
		},
	})
	if err != nil {
		log.Panicf("Error in batch write operation: %v", err)
	}

	fmt.Println("Batch seeding of products completed successfully!")
}
