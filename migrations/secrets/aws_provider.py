"""
AWS Secrets Manager based secrets provider.
"""

import json
from typing import Optional

from .exceptions import SecretAccessError, SecretNotFoundError
from .provider import SecretsProvider


class AwsSecretsProvider(SecretsProvider):
    """
    Secrets provider that reads from AWS Secrets Manager.
    Suitable for production AWS deployments.
    """

    def __init__(
        self,
        region_name: str = "us-east-1",
        secret_name: Optional[str] = None,
        endpoint_url: Optional[str] = None,
    ):
        """
        Initialize the AWS Secrets Manager provider.
        
        Args:
            region_name: AWS region name
            secret_name: The name of the secret in AWS Secrets Manager
                        (if storing multiple keys in one secret as JSON)
            endpoint_url: Optional custom endpoint URL (for LocalStack)
        """
        self._region_name = region_name
        self._secret_name = secret_name
        self._endpoint_url = endpoint_url
        self._client = None
        self._cached_secrets: Optional[dict] = None

    @property
    def provider_name(self) -> str:
        return "AWS Secrets Manager"

    def _get_client(self):
        """Lazily initialize the boto3 client."""
        if self._client is None:
            try:
                import boto3
            except ImportError:
                raise ImportError(
                    "boto3 is required for AWS Secrets Manager. "
                    "Install it with: pip install boto3"
                )
            
            kwargs = {"region_name": self._region_name}
            if self._endpoint_url:
                kwargs["endpoint_url"] = self._endpoint_url
            
            self._client = boto3.client("secretsmanager", **kwargs)
        return self._client

    def _load_secret_json(self) -> dict:
        """Load and cache the secret JSON if using a single secret name."""
        if self._cached_secrets is None and self._secret_name:
            client = self._get_client()
            response = client.get_secret_value(SecretId=self._secret_name)
            secret_string = response.get("SecretString", "{}")
            self._cached_secrets = json.loads(secret_string)
        return self._cached_secrets or {}

    def get_secret(self, key: str) -> str:
        """
        Retrieve a secret from AWS Secrets Manager.
        
        If secret_name was provided at init, looks up the key within that secret's JSON.
        Otherwise, treats the key as the secret name directly.
        
        Args:
            key: The secret key to retrieve
            
        Returns:
            The secret value
            
        Raises:
            SecretNotFoundError: If the secret is not found
            SecretAccessError: If there's an error accessing AWS
        """
        try:
            if self._secret_name:
                secrets = self._load_secret_json()
                value = secrets.get(key)
                if value is None:
                    raise SecretNotFoundError(key, self.provider_name)
                return value
            else:
                client = self._get_client()
                response = client.get_secret_value(SecretId=key)
                value = response.get("SecretString")
                if value is None:
                    raise SecretNotFoundError(key, self.provider_name)
                return value
        except SecretNotFoundError:
            raise
        except Exception as e:
            raise SecretAccessError(key, self.provider_name, e)
