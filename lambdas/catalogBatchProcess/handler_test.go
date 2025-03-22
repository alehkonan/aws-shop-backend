package main

import (
	"aws-shop-backend/packages/products"
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type MockDbClient struct {
	getProducts   func() (*dynamodb.ScanOutput, error)
	getProduct    func() (*dynamodb.GetItemOutput, error)
	createProduct func() (*dynamodb.TransactWriteItemsOutput, error)
}

func (client *MockDbClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	return client.getProducts()
}

func (client *MockDbClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return client.getProduct()
}

func (client *MockDbClient) TransactWriteItems(ctx context.Context, params *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error) {
	return client.createProduct()
}

type MockSnsClient struct{}

func (client *MockSnsClient) Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
	return &sns.PublishOutput{}, nil
}

func TestSuccessHandler(t *testing.T) {
	dbClient := &MockDbClient{
		createProduct: func() (*dynamodb.TransactWriteItemsOutput, error) {
			return &dynamodb.TransactWriteItemsOutput{}, nil
		},
	}
	productRepo = products.Repository(dbClient)
	snsClient = &MockSnsClient{}

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				Body: "product1,Description,12,10",
			},
		},
	}

	err := handler(context.Background(), sqsEvent)

	if err != nil {
		t.Errorf("error is not expected %v", err)
	}
}

func TestCreateProductError(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	dbClient := &MockDbClient{
		createProduct: func() (*dynamodb.TransactWriteItemsOutput, error) {
			return nil, errors.New("error")
		},
	}
	productRepo = products.Repository(dbClient)
	snsClient = &MockSnsClient{}

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				Body: "product1,Description,12,10",
			},
		},
	}

	handler(context.Background(), sqsEvent)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "can not create product") {
		t.Errorf("Expected log message not found. Got: %s", logOutput)
	}
}
