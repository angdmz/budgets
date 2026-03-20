package config

import (
	"fmt"
	"strings"
)

// SSLMode represents PostgreSQL SSL connection modes
type SSLMode string

const (
	SSLModeDisable    SSLMode = "disable"
	SSLModeAllow      SSLMode = "allow"
	SSLModePrefer     SSLMode = "prefer"
	SSLModeRequire    SSLMode = "require"
	SSLModeVerifyCA   SSLMode = "verify-ca"
	SSLModeVerifyFull SSLMode = "verify-full"
)

func (s SSLMode) String() string {
	return string(s)
}

func (s SSLMode) IsValid() bool {
	switch s {
	case SSLModeDisable, SSLModeAllow, SSLModePrefer, SSLModeRequire, SSLModeVerifyCA, SSLModeVerifyFull:
		return true
	}
	return false
}

func ParseSSLMode(value string) (SSLMode, error) {
	mode := SSLMode(strings.ToLower(strings.TrimSpace(value)))
	if !mode.IsValid() {
		return "", fmt.Errorf("invalid SSL mode: %s. Valid values: disable, allow, prefer, require, verify-ca, verify-full", value)
	}
	return mode, nil
}

// SecretString is a wrapper for sensitive string values that prevents accidental logging
type SecretString struct {
	value string
}

func NewSecretString(value string) SecretString {
	return SecretString{value: value}
}

func (s SecretString) Value() string {
	return s.value
}

func (s SecretString) String() string {
	if s.value == "" {
		return "[empty]"
	}
	return "[REDACTED]"
}

func (s SecretString) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED]"`), nil
}

func (s SecretString) IsEmpty() bool {
	return s.value == ""
}

func (s SecretString) GoString() string {
	return "[REDACTED]"
}

// Environment represents the application environment
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

func (e Environment) String() string {
	return string(e)
}

func (e Environment) IsProduction() bool {
	return e == EnvProduction
}

func (e Environment) IsDevelopment() bool {
	return e == EnvDevelopment
}

func ParseEnvironment(value string) Environment {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "production", "prod":
		return EnvProduction
	case "staging", "stage":
		return EnvStaging
	default:
		return EnvDevelopment
	}
}
