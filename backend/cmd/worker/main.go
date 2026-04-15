package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/aws"
	"github.com/HeadTDev/fitchallenge/internal/config"
)

const jobsQueueName = "fitchallenge-jobs"

type workerMessage struct {
	EventType string `json:"event_type"`
	UserID    string `json:"user_id"`
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

		for _, m := range msgs {
			if m.Body == nil {
				continue
			}

			processMessage(*m.Body)

			if m.ReceiptHandle != nil {
				if delErr := sqsClient.DeleteMessage(ctx, queueURL, *m.ReceiptHandle); delErr != nil {
					slog.Error("failed to delete processed SQS message", "error", delErr)
				}
			}
		}
	}
}

func processMessage(rawBody string) {
	var msg workerMessage
	if err := json.Unmarshal([]byte(rawBody), &msg); err != nil {
		slog.Warn("Unknown job type", "reason", "invalid_json", "body", rawBody)
		return
	}

	switch msg.EventType {
	case "log_submitted":
		slog.Info("Validating log for user", "user_id", msg.UserID)
	case "send_email":
		slog.Info("Processing send_email job", "user_id", msg.UserID)
	default:
		slog.Warn("Unknown job type", "event_type", msg.EventType)
	}
}
