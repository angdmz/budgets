"""
Secrets abstraction layer for the Budget Management System.
Provides pluggable secrets providers for different environments.
"""

from .exceptions import SecretAccessError, SecretNotFoundError
from .provider import SecretsProvider
from .env_provider import EnvSecretsProvider
from .aws_provider import AwsSecretsProvider
from .localstack_provider import LocalstackSecretsProvider
from .docker_provider import DockerSecretsProvider
from .factory import get_secrets_provider, SecretsProviderType

__all__ = [
    "SecretAccessError",
    "SecretNotFoundError",
    "SecretsProvider",
    "EnvSecretsProvider",
    "AwsSecretsProvider",
    "LocalstackSecretsProvider",
    "DockerSecretsProvider",
    "SecretsProviderType",
    "get_secrets_provider",
]
