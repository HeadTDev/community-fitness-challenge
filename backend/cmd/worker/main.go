package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/aws"
	"github.com/HeadTDev/fitchallenge/internal/config"
)

const jobsQueueName = "fitchallenge-jobs"
const notificationSender = "noreply@fitchallenge.local"

type workerMessage struct {
	SchemaVersion string `json:"schema_version"`
	Type          string `json:"type"`
	EventType     string `json:"event_type"`
	UserID        string `json:"user_id"`
	ChallengeID   string `json:"challenge_id"`
	To            string `json:"to"`
	Subject       string `json:"subject"`
	Body          string `json:"body"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.LoadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	awsCfg, err := aws.NewAWSConfig(ctx, cfg)
	if err != nil {
		slog.Error("failed to initialize AWS config", "error", err)
		os.Exit(1)
	}

	sqsClient := aws.NewSQSClient(awsCfg)
	sesClient := aws.NewSESClient(awsCfg)
	queueURL, err := sqsClient.GetQueueURL(ctx, jobsQueueName)
	if err != nil {
		slog.Error("failed to resolve worker queue URL", "queue", jobsQueueName, "error", err)
		os.Exit(1)
	}

	slog.Info("worker started", "queue", jobsQueueName, "queue_url", queueURL)

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker shutdown signal received, stopping after current cycle")
			return
		default:
		}

		msgs, recvErr := sqsClient.ReceiveMessages(ctx, queueURL, 5)
		if recvErr != nil {
			if ctx.Err() != nil {
				slog.Info("worker context cancelled while polling queue")
				return
			}
			slog.Error("failed to poll SQS queue", "error", recvErr)
			time.Sleep(2 * time.Second)
			continue
		}

		if len(msgs) == 0 {
			slog.Info("Worker heartbeat — queue empty")
			continue
		}

		for _, m := range msgs {
			if m.Body == nil {
				slog.Warn("received SQS message with empty body")
				continue
			}

			if processErr := processMessage(ctx, sesClient, *m.Body); processErr != nil {
				slog.Error("failed to process SQS message", "error", processErr)
				continue
			}

			if m.ReceiptHandle != nil {
				if delErr := sqsClient.DeleteMessage(ctx, queueURL, *m.ReceiptHandle); delErr != nil {
					slog.Error("failed to delete processed SQS message", "error", delErr)
				}
			}
		}
	}
}

func processMessage(ctx context.Context, sesClient *aws.SESClient, rawBody string) error {
	var msg workerMessage
	if err := json.Unmarshal([]byte(rawBody), &msg); err != nil {
		slog.Warn("Unknown job type", "reason", "invalid_json", "body", rawBody)
		return nil
	}

	eventType := strings.TrimSpace(msg.EventType)
	if eventType == "" {
		eventType = strings.TrimSpace(msg.Type)
	}

	switch eventType {
	case "log_submitted":
		slog.Info("Validating log for user", "user_id", msg.UserID)
		return nil
	case "send_email":
		to := strings.TrimSpace(msg.To)
		if to == "" {
			return fmt.Errorf("send_email job missing recipient")
		}
		subject := strings.TrimSpace(msg.Subject)
		if subject == "" {
			subject = "Community Fitness Challenge update"
		}
		body := strings.TrimSpace(msg.Body)
		if body == "" {
			body = "<p>You have a new notification from Community Fitness Challenge.</p>"
		}

		slog.Info("Sending email to user", "to", to, "user_id", msg.UserID, "challenge_id", msg.ChallengeID)
		msgID, err := sesClient.SendEmail(ctx, notificationSender, to, subject, body)
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
		slog.Info("Email sent", "to", to, "message_id", msgID)
		return nil
	default:
		slog.Warn("Unknown job type", "event_type", eventType)
		return nil
	}
}
