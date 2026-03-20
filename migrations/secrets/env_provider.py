"""
Environment variable based secrets provider.
"""

import os

from .exceptions import SecretNotFoundError
from .provider import SecretsProvider


class EnvSecretsProvider(SecretsProvider):
    """
    Secrets provider that reads from environment variables.
    Suitable for local development and simple deployments.
    """

    def __init__(self, prefix: str = ""):
        """
        Initialize the environment secrets provider.
        
        Args:
            prefix: Optional prefix to prepend to all secret keys
        """
        self._prefix = prefix

    @property
    def provider_name(self) -> str:
        return "environment variables"

    def get_secret(self, key: str) -> str:
        """
        Retrieve a secret from environment variables.
        
        Args:
            key: The environment variable name (prefix will be prepended)
            
        Returns:
            The environment variable value
            
        Raises:
            SecretNotFoundError: If the environment variable is not set
        """
        full_key = f"{self._prefix}{key}" if self._prefix else key
        value = os.environ.get(full_key)
        if value is None:
            raise SecretNotFoundError(full_key, self.provider_name)
        return value
