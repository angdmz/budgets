package secrets

const DefaultLocalstackEndpoint = "http://localhost:4566"

type LocalstackSecretsProvider struct {
	*AwsSecretsProvider
}

func NewLocalstackSecretsProvider(region, secretName, endpoint string) (*LocalstackSecretsProvider, error) {
	if endpoint == "" {
		endpoint = DefaultLocalstackEndpoint
	}

	awsProvider, err := NewAwsSecretsProvider(region, secretName, endpoint)
	if err != nil {
		return nil, err
	}

	return &LocalstackSecretsProvider{
		AwsSecretsProvider: awsProvider,
	}, nil
}
