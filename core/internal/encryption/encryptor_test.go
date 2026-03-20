package encryption

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)
	assert.NotEmpty(t, key)
	assert.Len(t, key, 44)
}

func TestEncryptor_EncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple text", "hello world"},
		{"empty string", ""},
		{"unicode", "こんにちは世界"},
		{"special chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"long text", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := encryptor.Encrypt(tt.plaintext)
			require.NoError(t, err)

			if tt.plaintext != "" {
				assert.NotEqual(t, tt.plaintext, encrypted)
			}

			decrypted, err := encryptor.Decrypt(encrypted)
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestEncryptor_EncryptDecryptMoney(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	tests := []struct {
		name     string
		amount   string
		currency string
	}{
		{"USD amount", "100.50", "USD"},
		{"EUR amount", "1234.56", "EUR"},
		{"zero amount", "0.00", "GBP"},
		{"large amount", "999999999.99", "JPY"},
		{"small decimals", "0.0001", "BTC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := decimal.NewFromString(tt.amount)
			require.NoError(t, err)

			money := Money{Amount: amount, Currency: tt.currency}

			encrypted, err := encryptor.EncryptMoney(money)
			require.NoError(t, err)
			assert.NotEmpty(t, encrypted)

			decrypted, err := encryptor.DecryptMoney(encrypted)
			require.NoError(t, err)
			assert.True(t, decrypted.Amount.Equal(amount))
			assert.Equal(t, tt.currency, decrypted.Currency)
		})
	}
}

func TestEncryptor_InvalidKey(t *testing.T) {
	_, err := NewEncryptor("invalid-key")
	assert.Error(t, err)
}

func TestEncryptor_DecryptInvalidCiphertext(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	_, err = encryptor.Decrypt("invalid-ciphertext")
	assert.Error(t, err)
}
