from pathlib import Path
from pydantic_settings import BaseSettings

SECRETS_DIR = Path("/run/secrets")

class Settings(BaseSettings):
    base_url: str = "http://localhost:8000"
    screenshots_dir: str = "/tests/screenshots"
    secrets_provider: str = "env"

    class Config:
        env_prefix = "INTEGRATION_TESTS_"
        extra = "ignore"

    @property
    def auth0_email(self) -> str:
        import os
        return os.getenv("INTEGRATION_TESTS_AUTH0_EMAIL", "")

    @property
    def auth0_password(self) -> str:
        if self.secrets_provider == "docker":
            return (SECRETS_DIR / "integration_tests_auth0_password").read_text().strip()
        import os
        return os.getenv("INTEGRATION_TESTS_AUTH0_PASSWORD", "")

settings = Settings()
