package encryption

import (
	"encoding/json"
	"fmt"

	"github.com/fernet/fernet-go"
	"github.com/shopspring/decimal"
)

type Encryptor struct {
	key *fernet.Key
}

func NewEncryptor(keyString string) (*Encryptor, error) {
	key, err := fernet.DecodeKey(keyString)
	if err != nil {
		return nil, fmt.Errorf("invalid fernet key: %w", err)
	}

	return &Encryptor{key: key}, nil
}

func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	token, err := fernet.EncryptAndSign([]byte(plaintext), e.key)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	return string(token), nil
}

func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	message := fernet.VerifyAndDecrypt([]byte(ciphertext), 0, []*fernet.Key{e.key})
	if message == nil {
		return "", fmt.Errorf("decryption failed: invalid token or key")
	}

	return string(message), nil
}

type Money struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
}

func NewMoney(amount decimal.Decimal, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

func (e *Encryptor) EncryptMoney(money Money) (string, error) {
	data := map[string]string{
		"amount":   money.Amount.String(),
		"currency": money.Currency,
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal money: %w", err)
	}

	return e.Encrypt(string(jsonBytes))
}

func (e *Encryptor) DecryptMoney(ciphertext string) (Money, error) {
	plaintext, err := e.Decrypt(ciphertext)
	if err != nil {
		return Money{}, err
	}

	var data map[string]string
	if err := json.Unmarshal([]byte(plaintext), &data); err != nil {
		return Money{}, fmt.Errorf("failed to unmarshal money: %w", err)
	}

	amount, err := decimal.NewFromString(data["amount"])
	if err != nil {
		return Money{}, fmt.Errorf("invalid amount: %w", err)
	}

	return Money{
		Amount:   amount,
		Currency: data["currency"],
	}, nil
}

func GenerateKey() (string, error) {
	key := fernet.Key{}
	if err := key.Generate(); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	return key.Encode(), nil
}
