package products

type Product struct {
	Id          string  `dynamodbav:"id"`
	Title       string  `dynamodbav:"title"`
	Description string  `dynamodbav:"description"`
	Price       float64 `dynamodbav:"price"`
}

type Stock struct {
	ProductId string `dynamodbav:"product_id"`
	Count     int    `dynamodbav:"count"`
}

type ProductDto struct {
	Id          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Count       int     `json:"count"`
}

type CreateProductDto struct {
	Title       string  `json:"title" validate:"required,min=1"`
	Description string  `json:"description" validate:"required,min=1"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Count       int     `json:"count" validate:"required,gte=0"`
}
