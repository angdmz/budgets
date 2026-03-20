package secrets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultDockerSecretsPath = "/run/secrets"

type DockerSecretsProvider struct {
	secretsPath string
}

func NewDockerSecretsProvider(secretsPath string) *DockerSecretsProvider {
	if secretsPath == "" {
		secretsPath = DefaultDockerSecretsPath
	}
	return &DockerSecretsProvider{secretsPath: secretsPath}
}

func (p *DockerSecretsProvider) GetSecret(key string) (string, error) {
	secretFile := filepath.Join(p.secretsPath, key)

	data, err := os.ReadFile(secretFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("secret '%s' not found in Docker secrets", key)
		}
		return "", fmt.Errorf("failed to read secret '%s': %w", key, err)
	}

	return strings.TrimSpace(string(data)), nil
}
