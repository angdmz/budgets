"""
Custom exceptions for secrets management.
"""


class SecretNotFoundError(Exception):
    """Raised when a required secret is not found."""

    def __init__(self, key: str, provider: str):
        self.key = key
        self.provider = provider
        super().__init__(f"Secret '{key}' not found in {provider}")


class SecretAccessError(Exception):
    """Raised when there's an error accessing the secrets provider."""

    def __init__(self, key: str, provider: str, cause: Exception):
        self.key = key
        self.provider = provider
        self.cause = cause
        super().__init__(f"Error accessing secret '{key}' from {provider}: {cause}")
