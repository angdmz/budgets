import os
from pathlib import Path

SECRETS_DIR = Path("/run/secrets")


class Settings:
    """Integration test settings, read from environment variables and Docker secrets."""

    def __init__(self):
        self.base_url: str = os.getenv("TEST_BASE_URL", "http://localhost:8000")
        self.screenshots_dir: str = os.getenv("SCREENSHOTS_DIR", "/tests/screenshots")
        self.secrets_provider: str = os.getenv("SECRETS_PROVIDER", "env")

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


settings = Settings()
