package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/budgets/core/internal/secrets"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Port int
	Env  Environment
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        SecretString
	Name            string
	SSLMode         SSLMode
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

type AuthConfig struct {
	Auth0Domain        string
	Auth0Audience      string
	Auth0ClientID      string
	Auth0ClientSecret  SecretString
	JWTSecret          SecretString
	EncryptionKey      SecretString
}

func (d DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password.Value(), d.Name, d.SSLMode.String(),
	)
}

func Load(provider secrets.SecretsProvider) (*Config, error) {
	encryptionKey, err := provider.GetSecret("encryption_key")
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	jwtSecret, err := provider.GetSecret("jwt_secret")
	if err != nil {
		return nil, fmt.Errorf("failed to get jwt secret: %w", err)
	}

	dbPassword, err := provider.GetSecret("db_password")
	if err != nil {
		return nil, fmt.Errorf("failed to get db password: %w", err)
	}

	auth0ClientSecret, err := provider.GetSecret("auth0_client_secret")
	if err != nil {
		return nil, fmt.Errorf("failed to get auth0 client secret: %w", err)
	}

	auth0Domain := getEnvOrDefault("AUTH0_DOMAIN", "")
	if auth0Domain == "" {
		return nil, fmt.Errorf("AUTH0_DOMAIN environment variable is required")
	}

	auth0Audience := getEnvOrDefault("AUTH0_AUDIENCE", "")
	if auth0Audience == "" {
		return nil, fmt.Errorf("AUTH0_AUDIENCE environment variable is required")
	}

	auth0ClientID := getEnvOrDefault("AUTH0_CLIENT_ID", "")
	if auth0ClientID == "" {
		return nil, fmt.Errorf("AUTH0_CLIENT_ID environment variable is required")
	}

	sslMode, err := ParseSSLMode(getEnvOrDefault("DB_SSLMODE", "disable"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_SSLMODE: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvOrDefaultInt("SERVER_PORT", 8080),
			Env:  ParseEnvironment(getEnvOrDefault("SERVER_ENV", "development")),
		},
		Database: DatabaseConfig{
			Host:            getEnvOrDefault("DB_HOSTNAME", "localhost"),
			Port:            getEnvOrDefaultInt("DB_PORT", 5432),
			User:            getEnvOrDefault("DB_USERNAME", "postgres"),
			Password:        NewSecretString(dbPassword),
			Name:            getEnvOrDefault("DB_NAME", "budgets"),
			SSLMode:         sslMode,
			MaxOpenConns:    getEnvOrDefaultInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvOrDefaultInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvOrDefaultInt("DB_CONN_MAX_LIFETIME_SECONDS", 300),
		},
		Auth: AuthConfig{
			Auth0Domain:       auth0Domain,
			Auth0Audience:     auth0Audience,
			Auth0ClientID:     auth0ClientID,
			Auth0ClientSecret: NewSecretString(auth0ClientSecret),
			JWTSecret:         NewSecretString(jwtSecret),
			EncryptionKey:     NewSecretString(encryptionKey),
		},
	}

	return cfg, nil
}

func getSecretOrDefault(provider secrets.SecretsProvider, key, defaultValue string) string {
	value, err := provider.GetSecret(key)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
