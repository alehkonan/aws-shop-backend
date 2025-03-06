package middleware

import (
	"context"
	"maps"

	"github.com/aws/aws-lambda-go/events"
)

func AddCorsHeaders() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			res, err := next(ctx, event)

			if res.Headers == nil {
				res.Headers = make(map[string]string)
			}

			corsHeaders := map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Headers":     "*",
				"Access-Control-Allow-Methods":     "*",
				"Access-Control-Allow-Credentials": "true",
			}

			maps.Copy(res.Headers, corsHeaders)

			return res, err
		}
	}
}
