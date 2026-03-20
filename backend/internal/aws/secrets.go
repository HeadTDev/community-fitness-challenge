package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type SecretsClient struct {
	client *secretsmanager.Client
}

func NewSecretsClient(cfg aws.Config) *SecretsClient {
	return &SecretsClient{
		client: secretsmanager.NewFromConfig(cfg),
	}
}

// GetSecret lekéri egy titok értékét a neve alapján.
func (s *SecretsClient) GetSecret(ctx context.Context, name string) (string, error) {
	output, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get secret %s: %w", name, err)
}
	if output.SecretString != nil {
		return *output.SecretString, nil
	}
	return "", fmt.Errorf("secret %s has no string value", name)
}

// CreateSecret létrehoz egy új titkot.
func (s *SecretsClient) CreateSecret(ctx context.Context, name, value string) error {
	_, err := s.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(name),
		SecretString: aws.String(value),
	})
	if err != nil {
		return fmt.Errorf("failed to create secret %s: %w", name, err)
	}
	return nil
}
