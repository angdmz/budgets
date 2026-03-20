"""
LocalStack based secrets provider for development.
"""

from typing import Optional

from .aws_provider import AwsSecretsProvider


class LocalstackSecretsProvider(AwsSecretsProvider):
    """
    Secrets provider that reads from LocalStack's Secrets Manager.
    Suitable for local development with AWS-like infrastructure.
    """

    DEFAULT_ENDPOINT = "http://localhost:4566"

    def __init__(
        self,
        region_name: str = "us-east-1",
        secret_name: Optional[str] = None,
        endpoint_url: str = DEFAULT_ENDPOINT,
    ):
        """
        Initialize the LocalStack Secrets Manager provider.
        
        Args:
            region_name: AWS region name (default: us-east-1)
            secret_name: The name of the secret in LocalStack Secrets Manager
            endpoint_url: LocalStack endpoint URL (default: http://localhost:4566)
        """
        super().__init__(
            region_name=region_name,
            secret_name=secret_name,
            endpoint_url=endpoint_url,
        )

    @property
    def provider_name(self) -> str:
        return "LocalStack Secrets Manager"
