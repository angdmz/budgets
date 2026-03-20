"""
Docker secrets based secrets provider.
Reads secrets from /run/secrets/<secret_name> as mounted by Docker Swarm or docker-compose.
"""

from pathlib import Path

from .exceptions import SecretAccessError, SecretNotFoundError
from .provider import SecretsProvider


class DockerSecretsProvider(SecretsProvider):
    """
    Secrets provider that reads from Docker secrets.
    Docker secrets are mounted as files in /run/secrets/ directory.
    Suitable for Docker Swarm and docker-compose deployments.
    """

    DEFAULT_SECRETS_PATH = "/run/secrets"

    def __init__(self, secrets_path: str = DEFAULT_SECRETS_PATH):
        """
        Initialize the Docker secrets provider.
        
        Args:
            secrets_path: Path to the secrets directory (default: /run/secrets)
        """
        self._secrets_path = Path(secrets_path)

    @property
    def provider_name(self) -> str:
        return "Docker secrets"

    def get_secret(self, key: str) -> str:
        """
        Retrieve a secret from Docker secrets.
        
        Args:
            key: The secret name (file name in /run/secrets/)
            
        Returns:
            The secret value (file contents, stripped of whitespace)
            
        Raises:
            SecretNotFoundError: If the secret file does not exist
            SecretAccessError: If there's an error reading the file
        """
        secret_file = self._secrets_path / key
        
        if not secret_file.exists():
            raise SecretNotFoundError(key, self.provider_name)
        
        try:
            return secret_file.read_text().strip()
        except Exception as e:
            raise SecretAccessError(key, self.provider_name, e)
