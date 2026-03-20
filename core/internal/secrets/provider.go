package secrets

type SecretsProvider interface {
	GetSecret(key string) (string, error)
}
