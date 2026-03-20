package secrets

import (
	"os"
	"strings"
)

const (
	ProviderTypeEnv        = "env"
	ProviderTypeAws        = "aws"
	ProviderTypeLocalstack = "localstack"
	ProviderTypeDocker     = "docker"
)

func GetProvider() SecretsProvider {
	providerType := strings.ToLower(os.Getenv("SECRETS_PROVIDER"))
	if providerType == "" {
		providerType = ProviderTypeEnv
	}

	switch providerType {
	case ProviderTypeAws:
		region := os.Getenv("AWS_REGION")
		if region == "" {
			region = "us-east-1"
		}
		secretName := os.Getenv("AWS_SECRET_NAME")
		provider, err := NewAwsSecretsProvider(region, secretName, "")
		if err != nil {
			return NewEnvSecretsProvider("")
		}
		return provider

	case ProviderTypeLocalstack:
		region := os.Getenv("AWS_REGION")
		if region == "" {
			region = "us-east-1"
		}
		secretName := os.Getenv("AWS_SECRET_NAME")
		endpoint := os.Getenv("LOCALSTACK_ENDPOINT")
		provider, err := NewLocalstackSecretsProvider(region, secretName, endpoint)
		if err != nil {
			return NewEnvSecretsProvider("")
		}
		return provider

	case ProviderTypeDocker:
		secretsPath := os.Getenv("DOCKER_SECRETS_PATH")
		return NewDockerSecretsProvider(secretsPath)

	default:
		prefix := os.Getenv("SECRETS_PREFIX")
		return NewEnvSecretsProvider(prefix)
	}
}
