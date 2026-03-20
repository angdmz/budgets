package secrets

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvSecretsProvider_GetSecret(t *testing.T) {
	provider := NewEnvSecretsProvider("")

	t.Run("existing secret", func(t *testing.T) {
		os.Setenv("TEST_SECRET_KEY", "test-value")
		defer os.Unsetenv("TEST_SECRET_KEY")

		value, err := provider.GetSecret("TEST_SECRET_KEY")
		require.NoError(t, err)
		assert.Equal(t, "test-value", value)
	})

	t.Run("missing secret", func(t *testing.T) {
		_, err := provider.GetSecret("NON_EXISTENT_KEY")
		assert.Error(t, err)
	})
}

func TestEnvSecretsProvider_WithPrefix(t *testing.T) {
	provider := NewEnvSecretsProvider("APP_")

	os.Setenv("APP_DATABASE_URL", "postgres://localhost/test")
	defer os.Unsetenv("APP_DATABASE_URL")

	value, err := provider.GetSecret("DATABASE_URL")
	require.NoError(t, err)
	assert.Equal(t, "postgres://localhost/test", value)
}
