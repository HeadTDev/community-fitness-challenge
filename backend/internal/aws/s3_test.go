package aws_test

import (
	"context"
	"os"
	"strings"
	"testing"

	fca "github.com/HeadTDev/fitchallenge/internal/aws"
	"github.com/HeadTDev/fitchallenge/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3Client(t *testing.T) {
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("Skipping integration test: AWS_ENDPOINT_URL not set")
	}

	ctx := context.Background()
	cfg := config.LoadConfig()

	awsCfg, err := fca.NewAWSConfig(ctx, cfg)
	require.NoError(t, err)

	s3Client := fca.NewS3Client(awsCfg, cfg.AWSEndpoint)
	bucket := "fitchallenge-assets"
	testKey := "tests/test-file.txt"
	testContent := "hello world"

	t.Run("Upload File", func(t *testing.T) {
		body := strings.NewReader(testContent)
		url, err := s3Client.UploadFile(ctx, bucket, testKey, body, "text/plain")
		assert.NoError(t, err)
		assert.Contains(t, url, bucket)
		assert.Contains(t, url, testKey)
	})

	t.Run("List Files", func(t *testing.T) {
		files, err := s3Client.ListFiles(ctx, bucket, "tests/")
		assert.NoError(t, err)
		assert.Contains(t, files, testKey)
	})

	t.Run("Delete File", func(t *testing.T) {
		err := s3Client.DeleteFile(ctx, bucket, testKey)
		assert.NoError(t, err)

		files, err := s3Client.ListFiles(ctx, bucket, "tests/")
		assert.NoError(t, err)
		assert.NotContains(t, files, testKey)
	})
}
