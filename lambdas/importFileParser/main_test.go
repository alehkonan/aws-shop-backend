package main

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockS3Client struct{}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader("uploaded/test.csv")),
	}, nil
}

func (m *mockS3Client) CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
	return &s3.CopyObjectOutput{}, nil
}

func (m *mockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return &s3.DeleteObjectOutput{}, nil
}

func TestParseCsv(t *testing.T) {
	tests := []struct {
		name        string
		csvData     string
		expectError bool
	}{
		{
			name:        "Valid CSV",
			csvData:     "col1,col2\nval1,val2",
			expectError: false,
		},
		{
			name:        "Invalid CSV",
			csvData:     "col1,col2\nval1,val2,val3",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &s3.GetObjectOutput{
				Body: io.NopCloser(strings.NewReader(tt.csvData)),
			}

			err := parseCsv(obj)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	s3Client = &mockS3Client{}
	s3Event := events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{
						Name: "test-bucket",
					},
					Object: events.S3Object{
						Key: "uploaded/test.csv",
					},
				},
			},
		},
	}

	err := handler(context.Background(), s3Event)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
