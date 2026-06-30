import os
from pathlib import Path
from pydantic_settings import BaseSettings
from pydantic import ConfigDict

SECRETS_DIR = Path("/run/secrets")


class Settings(BaseSettings):
    """Integration test settings, read from environment variables and Docker secrets."""

    base_url: str = "http://localhost:8000"
    screenshots_dir: str = "/tests/screenshots"
    secrets_provider: str = "env"

    model_config = ConfigDict(env_prefix="INTEGRATION_TESTS_", extra="ignore")

    @property
    def auth0_domain(self) -> str:
        return os.getenv("AUTH0_DOMAIN", "")

    @property
    def auth0_mgmt_client_id(self) -> str:
        return os.getenv("AUTH0_MGMT_CLIENT_ID", "")

    @property
    def auth0_mgmt_client_secret(self) -> str:
        if self.secrets_provider == "docker":
            try:
                return (SECRETS_DIR / "auth0_mgmt_client_secret").read_text().strip()
            except FileNotFoundError:
                return ""
        return os.getenv("AUTH0_MGMT_CLIENT_SECRET", "")

    @property
    def auth0_client_id(self) -> str:
        return os.getenv("AUTH0_CLIENT_ID", "")

    @property
    def auth0_client_secret(self) -> str:
        if self.secrets_provider == "docker":
            try:
                return (SECRETS_DIR / "auth0_client_secret").read_text().strip()
            except FileNotFoundError:
                return ""
        return os.getenv("AUTH0_CLIENT_SECRET", "")

    @property
    def auth0_audience(self) -> str:
        return os.getenv("AUTH0_AUDIENCE", "https://api.budget.local")

    @property
    def auth0_db_connection(self) -> str:
        return os.getenv("AUTH0_DB_CONNECTION", "Username-Password-Authentication")

    @property
    def auth0_email(self) -> str:
        if self.secrets_provider == "docker":
            return (SECRETS_DIR / "integration_tests_auth0_email").read_text().strip()
        return os.getenv("INTEGRATION_TESTS_AUTH0_EMAIL", "")

    @property
    def auth0_password(self) -> str:
        if self.secrets_provider == "docker":
            return (SECRETS_DIR / "integration_tests_auth0_password").read_text().strip()
        return os.getenv("INTEGRATION_TESTS_AUTH0_PASSWORD", "")


settings = Settings()
