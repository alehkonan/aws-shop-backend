package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	username string
	password string
)

func init() {
	username = os.Getenv("USERNAME")
	password = os.Getenv("PASSWORD")
	if username == "" || password == "" {
		log.Fatal("no env variables found")
	}
}

func createPolicy(effect string, resource string) events.APIGatewayCustomAuthorizerPolicy {
	return events.APIGatewayCustomAuthorizerPolicy{
		Version: "2012-10-17",
		Statement: []events.IAMPolicyStatement{
			{
				Action:   []string{"execute-api:Invoke"},
				Effect:   effect,
				Resource: []string{resource},
			},
		},
	}
}

func handler(
	event events.APIGatewayCustomAuthorizerRequest,
) (events.APIGatewayCustomAuthorizerResponse, error) {
	res := events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: username,
	}

	if err := validateToken(event.AuthorizationToken); err != nil {
		log.Printf("token validation error: %v", err)
		res.PolicyDocument = createPolicy("Deny", event.MethodArn)
		return res, nil
	}

	res.PolicyDocument = createPolicy("Allow", event.MethodArn)
	return res, nil
}

func main() {
	lambda.Start(handler)
}
