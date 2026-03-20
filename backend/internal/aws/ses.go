package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type SESClient struct {
	client *ses.Client
}

func NewSESClient(cfg aws.Config) *SESClient {
	return &SESClient{
		client: ses.NewFromConfig(cfg),
	}
}

// SendEmail elküld egy HTML alapú e-mailt a megadott címzettnek.
func (s *SESClient) SendEmail(ctx context.Context, from, to, subject, htmlBody string) (string, error) {
	output, err := s.client.SendEmail(ctx, &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Data: aws.String(htmlBody),
				},
			},
			Subject: &types.Content{
				Data: aws.String(subject),
			},
		},
		Source: aws.String(from),
	})
	if err != nil {
		return "", fmt.Errorf("failed to send email from %s to %s: %w", from, to, err)
	}
	return *output.MessageId, nil
}
