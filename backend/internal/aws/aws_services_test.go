package aws_test

import (
	"context"
	"os"
	"testing"

	fca "github.com/HeadTDev/fitchallenge/internal/aws"
	"github.com/HeadTDev/fitchallenge/internal/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAWSServices(t *testing.T) {
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("Skipping integration test: AWS_ENDPOINT_URL not set")
	}

	ctx := context.Background()
	cfg := config.LoadConfig()

	awsCfg, err := fca.NewAWSConfig(ctx, cfg)
	require.NoError(t, err)

	t.Run("SQS Operations", func(t *testing.T) {
		sqsClient := fca.NewSQSClient(awsCfg)
		queueName := "fitchallenge-jobs"

		// Megjegyzés: A sor már létezik az init-aws.sh miatt
		queueURL, err := sqsClient.GetQueueURL(ctx, queueName)
		require.NoError(t, err)
		assert.NotEmpty(t, queueURL)

		// Send
		msgBody := "test message content"
		msgID, err := sqsClient.SendMessage(ctx, queueURL, msgBody)
		assert.NoError(t, err)
		assert.NotEmpty(t, msgID)

		// Receive
		msgs, err := sqsClient.ReceiveMessages(ctx, queueURL, 1)
		assert.NoError(t, err)
		assert.Len(t, msgs, 1)
		assert.Equal(t, msgBody, *msgs[0].Body)

		// Delete
		err = sqsClient.DeleteMessage(ctx, queueURL, *msgs[0].ReceiptHandle)
		assert.NoError(t, err)
	})

	t.Run("SES Operations", func(t *testing.T) {
		sesClient := fca.NewSESClient(awsCfg)
		from := "noreply@fitchallenge.local"
		to := "user@example.com"
		subject := "Test Subject"
		body := "<h1>Hello</h1>"

		msgID, err := sesClient.SendEmail(ctx, from, to, subject, body)
		assert.NoError(t, err)
		assert.NotEmpty(t, msgID)
	})

	t.Run("Secrets Manager Operations", func(t *testing.T) {
		secretsClient := fca.NewSecretsClient(awsCfg)
		secretName := "test/secret-" + uuid.New().String()
		secretValue := "super-secret-value"

		// Create
		err := secretsClient.CreateSecret(ctx, secretName, secretValue)
		assert.NoError(t, err)

		// Get
		val, err := secretsClient.GetSecret(ctx, secretName)
		assert.NoError(t, err)
		assert.Equal(t, secretValue, val)
	})
}
