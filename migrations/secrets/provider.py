"""
Abstract base class for secrets providers.
"""

from abc import ABC, abstractmethod

from .exceptions import SecretNotFoundError


class SecretsProvider(ABC):
    """
    Abstract interface for secrets retrieval.
    All concrete implementations must implement get_secret method.
    """

    @property
    @abstractmethod
    def provider_name(self) -> str:
        """Return the name of this provider for error messages."""
        pass

    @abstractmethod
    def get_secret(self, key: str) -> str:
        """
        Retrieve a secret value by its key.
        
        Args:
            key: The secret key/name to retrieve
            
        Returns:
            The secret value
            
        Raises:
            SecretNotFoundError: If the secret is not found
            SecretAccessError: If there's an error accessing the provider
        """
        pass
