package middleware

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type Handler func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type Middleware func(Handler) Handler
