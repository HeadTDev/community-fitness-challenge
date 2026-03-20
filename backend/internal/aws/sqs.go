package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSClient struct {
	client *sqs.Client
}

func NewSQSClient(cfg aws.Config) *SQSClient {
	return &SQSClient{
		client: sqs.NewFromConfig(cfg),
	}
}

// GetQueueURL lekéri egy sor URL-jét a neve alapján.
func (s *SQSClient) GetQueueURL(ctx context.Context, queueName string) (string, error) {
	output, err := s.client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get queue URL for %s: %w", queueName, err)
	}
	return *output.QueueUrl, nil
}

// SendMessage elküld egy üzenetet a megadott sorba.
func (s *SQSClient) SendMessage(ctx context.Context, queueURL, body string) (string, error) {
	output, err := s.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(body),
	})
	if err != nil {
		return "", fmt.Errorf("failed to send message to %s: %w", queueURL, err)
	}
	return *output.MessageId, nil
}

// ReceiveMessages beolvas üzeneteket a sorból (max megadott számút).
func (s *SQSClient) ReceiveMessages(ctx context.Context, queueURL string, maxMessages int32) ([]types.Message, error) {
	output, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: maxMessages,
		WaitTimeSeconds:     10, // Long polling
	})
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages from %s: %w", queueURL, err)
	}
	return output.Messages, nil
}

// DeleteMessage töröl egy üzenetet a sorból a receiptHandle alapján.
func (s *SQSClient) DeleteMessage(ctx context.Context, queueURL, receiptHandle string) error {
	_, err := s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		return fmt.Errorf("failed to delete message from %s: %w", queueURL, err)
	}
	return nil
}
