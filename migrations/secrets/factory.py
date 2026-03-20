"""
Factory for creating secrets providers based on configuration.
"""

import os
from typing import Optional

from .provider import SecretsProvider
from .env_provider import EnvSecretsProvider
from .aws_provider import AwsSecretsProvider
from .localstack_provider import LocalstackSecretsProvider
from .docker_provider import DockerSecretsProvider


class SecretsProviderType:
    """Enumeration of supported secrets provider types."""
    ENV = "env"
    AWS = "aws"
    LOCALSTACK = "localstack"
    DOCKER = "docker"


def get_secrets_provider(
    provider_type: Optional[str] = None,
    **kwargs,
) -> SecretsProvider:
    """
    Factory function to create a secrets provider based on configuration.
    
    The provider type can be specified directly or via the SECRETS_PROVIDER
    environment variable.
    
    Args:
        provider_type: The type of provider to create (env, aws, localstack, docker)
        **kwargs: Additional arguments passed to the provider constructor
        
    Returns:
        A configured SecretsProvider instance
        
    Raises:
        ValueError: If an unknown provider type is specified
    """
    if provider_type is None:
        provider_type = os.environ.get("SECRETS_PROVIDER", SecretsProviderType.ENV)
    
    provider_type = provider_type.lower()
    
    if provider_type == SecretsProviderType.ENV:
        prefix = kwargs.get("prefix", os.environ.get("SECRETS_PREFIX", ""))
        return EnvSecretsProvider(prefix=prefix)
    
    elif provider_type == SecretsProviderType.AWS:
        return AwsSecretsProvider(
            region_name=kwargs.get("region_name", os.environ.get("AWS_REGION", "us-east-1")),
            secret_name=kwargs.get("secret_name", os.environ.get("AWS_SECRET_NAME")),
            endpoint_url=kwargs.get("endpoint_url"),
        )
    
    elif provider_type == SecretsProviderType.LOCALSTACK:
        return LocalstackSecretsProvider(
            region_name=kwargs.get("region_name", os.environ.get("AWS_REGION", "us-east-1")),
            secret_name=kwargs.get("secret_name", os.environ.get("AWS_SECRET_NAME")),
            endpoint_url=kwargs.get(
                "endpoint_url",
                os.environ.get("LOCALSTACK_ENDPOINT", LocalstackSecretsProvider.DEFAULT_ENDPOINT),
            ),
        )
    
    elif provider_type == SecretsProviderType.DOCKER:
        return DockerSecretsProvider(
            secrets_path=kwargs.get(
                "secrets_path",
                os.environ.get("DOCKER_SECRETS_PATH", DockerSecretsProvider.DEFAULT_SECRETS_PATH),
            ),
        )
    
    else:
        raise ValueError(
            f"Unknown secrets provider type: '{provider_type}'. "
            f"Supported types: {SecretsProviderType.ENV}, {SecretsProviderType.AWS}, "
            f"{SecretsProviderType.LOCALSTACK}, {SecretsProviderType.DOCKER}"
        )
