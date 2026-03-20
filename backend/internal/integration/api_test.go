package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/adapter/postgres"
	fca "github.com/HeadTDev/fitchallenge/internal/aws"
	"github.com/HeadTDev/fitchallenge/internal/config"
	handler "github.com/HeadTDev/fitchallenge/internal/handler/http"
	"github.com/HeadTDev/fitchallenge/internal/handler/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullIntegration(t *testing.T) {
	if os.Getenv("DB_HOST") == "" || os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("Skipping integration test: DB_HOST or AWS_ENDPOINT_URL not set")
	}

	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := config.LoadConfig()

	// 1. Test Database Pool
	pool, err := postgres.NewConnection(ctx, cfg)
	require.NoError(t, err)
	defer pool.Close()

	// 2. Test AWS / S3 Client
	awsCfg, err := fca.NewAWSConfig(ctx, cfg)
	require.NoError(t, err)
	s3Client := fca.NewS3Client(awsCfg, cfg.S3PublicURL)

	// 3. Test API Handlers
	h := handler.NewHealthHandler(pool)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.GET("/readyz", h.Readyz)

	t.Run("API Readyz Check", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/readyz", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		
		assert.Equal(t, true, resp["success"])

		// Meta check
		meta := resp["meta"].(map[string]interface{})
		assert.NotEmpty(t, meta["request_id"])

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "ready", data["status"])
		assert.Equal(t, "ok", data["db"])
	})

	t.Run("S3 Integration", func(t *testing.T) {
		bucket := "fitchallenge-assets"
		key := "integration-test/hello.txt"
		content := "integration test content"
		
		// Upload
		url, err := s3Client.UploadFile(ctx, bucket, key, strings.NewReader(content), "text/plain")
		assert.NoError(t, err)
		assert.Contains(t, url, cfg.S3PublicURL)
		assert.Contains(t, url, bucket)
		assert.Contains(t, url, key)

		// List
		files, err := s3Client.ListFiles(ctx, bucket, "integration-test/")
		assert.NoError(t, err)
		assert.Contains(t, files, key)

		// Delete
		err = s3Client.DeleteFile(ctx, bucket, key)
		assert.NoError(t, err)
	})
}
