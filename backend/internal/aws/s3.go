package aws

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	client   *s3.Client
	endpoint string
}

func NewS3Client(cfg aws.Config, endpoint string) *S3Client {
	return &S3Client{
		client:   s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.UsePathStyle = true // LocalStack-hez elengedhetetlen a PathStyle
		}),
		endpoint: endpoint,
	}
}

// UploadFile feltölt egy fájlt az S3-ba és visszaadja a publikus URL-t.
func (s *S3Client) UploadFile(ctx context.Context, bucket, key string, body io.Reader, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return s.GetFileURL(bucket, key), nil
}

// GetFileURL visszaadja a fájl elérési útját.
func (s *S3Client) GetFileURL(bucket, key string) string {
	if s.endpoint != "" {
		// LocalStack esetén: http://localhost:4566/bucket/key
		endpoint := strings.Replace(s.endpoint, "localstack", "localhost", 1)
		return fmt.Sprintf("%s/%s/%s", endpoint, bucket, key)
	}
	// Valódi AWS esetén: https://bucket.s3.region.amazonaws.com/key
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)
}

// ListFiles listázza a fájlokat egy prefix alapján.
func (s *S3Client) ListFiles(ctx context.Context, bucket, prefix string) ([]string, error) {
	output, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files from S3: %w", err)
	}

	var files []string
	for _, obj := range output.Contents {
		files = append(files, *obj.Key)
	}
	return files, nil
}

// DeleteFile töröl egy fájlt az S3-ból.
func (s *S3Client) DeleteFile(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}
	return nil
}
