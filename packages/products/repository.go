package products

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type DbClient interface {
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	TransactWriteItems(ctx context.Context, params *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
}

type ProductRepository struct {
	dbClient      DbClient
	productsTable string
	stocksTable   string
}

func Repository(dbClient DbClient) *ProductRepository {
	return &ProductRepository{
		dbClient:      dbClient,
		productsTable: os.Getenv("PRODUCTS_TABLE"),
		stocksTable:   os.Getenv("STOCKS_TABLE"),
	}
}

func (r *ProductRepository) GetAllProducts(ctx context.Context) ([]ProductDto, error) {
	productsResult, err := r.dbClient.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.productsTable),
	})
	if err != nil {
		return nil, err
	}

	stocksResult, err := r.dbClient.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.stocksTable),
	})
	if err != nil {
		return nil, err
	}

	var products []Product
	err = attributevalue.UnmarshalListOfMaps(productsResult.Items, &products)
	if err != nil {
		return nil, err
	}

	var stocks []Stock
	err = attributevalue.UnmarshalListOfMaps(stocksResult.Items, &stocks)
	if err != nil {
		return nil, err
	}

	stockMap := make(map[string]int)
	for _, stock := range stocks {
		stockMap[stock.ProductId] = stock.Count
	}

	result := make([]ProductDto, 0, len(products))
	for _, product := range products {
		count := stockMap[product.Id]

		result = append(result, ProductDto{
			Id:          product.Id,
			Title:       product.Title,
			Description: product.Description,
			Price:       product.Price,
			Image:       product.Image,
			Count:       count,
		})
	}

	return result, nil
}

func (r *ProductRepository) GetProductById(ctx context.Context, productId string) (*ProductDto, error) {
	productResult, err := r.dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.productsTable),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{
				Value: productId,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if productResult.Item == nil {
		return nil, nil
	}

	stockResult, err := r.dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.stocksTable),
		Key: map[string]types.AttributeValue{
			"product_id": &types.AttributeValueMemberS{
				Value: productId,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var product Product
	err = attributevalue.UnmarshalMap(productResult.Item, &product)
	if err != nil {
		return nil, err
	}

	var stock Stock
	if stockResult.Item != nil {
		err = attributevalue.UnmarshalMap(stockResult.Item, &stock)
		if err != nil {
			return nil, err
		}
	}

	return &ProductDto{
		Id:          product.Id,
		Title:       product.Title,
		Description: product.Description,
		Price:       product.Price,
		Image:       product.Image,
		Count:       stock.Count,
	}, nil
}

func (r *ProductRepository) CreateProduct(ctx context.Context, dto CreateProductDto) (*ProductDto, error) {
	product := ProductDto{
		Id:          uuid.New().String(),
		Title:       dto.Title,
		Description: dto.Description,
		Price:       dto.Price,
		Count:       dto.Count,
		Image:       dto.Image,
	}

	productItem, err := attributevalue.MarshalMap(Product{
		Id:          product.Id,
		Title:       product.Title,
		Description: product.Description,
		Price:       product.Price,
		Image:       product.Image,
	})
	if err != nil {
		return nil, err
	}

	stockItem, err := attributevalue.MarshalMap(Stock{
		ProductId: product.Id,
		Count:     product.Count,
	})
	if err != nil {
		return nil, err
	}

	_, err = r.dbClient.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Put: &types.Put{
					TableName: aws.String(r.productsTable),
					Item:      productItem,
				},
			},
			{
				Put: &types.Put{
					TableName: aws.String(r.stocksTable),
					Item:      stockItem,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &product, nil
}
