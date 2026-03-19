package aws

import (
	"context"
	"log"

	projectConfig "github.com/HeadTDev/fitchallenge/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// NewAWSConfig létrehozza az AWS konfigurációt, figyelembe véve a LocalStack endpointot.
func NewAWSConfig(ctx context.Context, cfg *projectConfig.Config) (aws.Config, error) {
	var opts []func(*config.LoadOptions) error

	// Alapértelmezett régió beállítása
	opts = append(opts, config.WithRegion(cfg.AWSRegion))

	// Ha van megadva AWSEndpoint, akkor LocalStack-et használunk (fejlesztés)
	if cfg.AWSEndpoint != "" {
		log.Printf("☁️ Using LocalStack endpoint: %s", cfg.AWSEndpoint)
		
		// Senior tipp: LocalStack-hez fix teszt hitelesítő adatok kellenek.
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		))

		// Endpoint felüldefiniálása
		opts = append(opts, config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           cfg.AWSEndpoint,
					SigningRegion: cfg.AWSRegion,
				}, nil
			}),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, err
	}

	return awsCfg, nil
}
