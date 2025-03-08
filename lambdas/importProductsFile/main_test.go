package main

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

type testCase struct {
	name         string
	fileName     string
	expectError  bool
	expectStatus int
}

func TestImportProductsHandler(t *testing.T) {
	cases := []testCase{
		{
			name:         "missing file name",
			fileName:     "",
			expectError:  false,
			expectStatus: 400,
		},
		{
			name:         "success",
			fileName:     "test.csv",
			expectError:  false,
			expectStatus: 200,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("BUCKET_NAME", "test_bucket")

			req := events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"name": tt.fileName,
				},
			}

			res, err := handleRequest(context.Background(), req)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectStatus != res.StatusCode {
				t.Errorf("expected status code %d, got %d", tt.expectStatus, res.StatusCode)
			}
		})
	}
}
