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
	GoogleClientID     string
	GoogleClientSecret SecretString
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
	encryptionKey, err := provider.GetSecret("ENCRYPTION_KEY")
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	jwtSecret := getSecretOrDefault(provider, "JWT_SECRET", "default-jwt-secret-change-in-production")
	dbPassword := getSecretOrDefault(provider, "DB_PASSWORD", "postgres")
	googleClientID := getSecretOrDefault(provider, "GOOGLE_CLIENT_ID", "")
	googleClientSecret := getSecretOrDefault(provider, "GOOGLE_CLIENT_SECRET", "")

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
			GoogleClientID:     googleClientID,
			GoogleClientSecret: NewSecretString(googleClientSecret),
			JWTSecret:          NewSecretString(jwtSecret),
			EncryptionKey:      NewSecretString(encryptionKey),
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
