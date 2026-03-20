package secrets

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type AwsSecretsProvider struct {
	client        *secretsmanager.SecretsManager
	secretName    string
	cachedSecrets map[string]string
}

func NewAwsSecretsProvider(region, secretName, endpoint string) (*AwsSecretsProvider, error) {
	awsConfig := &aws.Config{
		Region: aws.String(region),
	}

	if endpoint != "" {
		awsConfig.Endpoint = aws.String(endpoint)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	client := secretsmanager.New(sess)

	return &AwsSecretsProvider{
		client:        client,
		secretName:    secretName,
		cachedSecrets: nil,
	}, nil
}

func (p *AwsSecretsProvider) loadSecrets() error {
	if p.cachedSecrets != nil {
		return nil
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(p.secretName),
	}

	result, err := p.client.GetSecretValue(input)
	if err != nil {
		return fmt.Errorf("failed to get secret value: %w", err)
	}

	if result.SecretString == nil {
		return fmt.Errorf("secret string is nil")
	}

	p.cachedSecrets = make(map[string]string)
	if err := json.Unmarshal([]byte(*result.SecretString), &p.cachedSecrets); err != nil {
		p.cachedSecrets[p.secretName] = *result.SecretString
	}

	return nil
}

func (p *AwsSecretsProvider) GetSecret(key string) (string, error) {
	if p.secretName != "" {
		if err := p.loadSecrets(); err != nil {
			return "", err
		}

		value, ok := p.cachedSecrets[key]
		if !ok {
			return "", fmt.Errorf("secret %s not found", key)
		}
		return value, nil
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	}

	result, err := p.client.GetSecretValue(input)
	if err != nil {
		return "", fmt.Errorf("failed to get secret %s: %w", key, err)
	}

	if result.SecretString == nil {
		return "", fmt.Errorf("secret %s has no string value", key)
	}

	return *result.SecretString, nil
}
