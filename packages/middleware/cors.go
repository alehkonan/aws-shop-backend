package middleware

import (
	"context"
	"maps"

	"github.com/aws/aws-lambda-go/events"
)

func AddCorsHeaders() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
			res := next(ctx, event)

			if res.Headers == nil {
				res.Headers = make(map[string]string)
			}

			corsHeaders := map[string]string{
				res.Headers["Access-Control-Allow-Origin"]:      "*",
				res.Headers["Access-Control-Allow-Headers"]:     "*",
				res.Headers["Access-Control-Allow-Methods"]:     "*",
				res.Headers["Access-Control-Allow-Credentials"]: "true",
			}

			maps.Copy(res.Headers, corsHeaders)

			return res
		}
	}
}
