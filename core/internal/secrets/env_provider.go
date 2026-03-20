package secrets

import (
	"fmt"
	"os"
)

type EnvSecretsProvider struct {
	prefix string
}

func NewEnvSecretsProvider(prefix string) *EnvSecretsProvider {
	return &EnvSecretsProvider{prefix: prefix}
}

func (p *EnvSecretsProvider) GetSecret(key string) (string, error) {
	fullKey := key
	if p.prefix != "" {
		fullKey = p.prefix + key
	}

	value := os.Getenv(fullKey)
	if value == "" {
		return "", fmt.Errorf("secret %s not found in environment", fullKey)
	}

	return value, nil
}
