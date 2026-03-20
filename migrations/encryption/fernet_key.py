"""
Pydantic type for Fernet encryption keys.
"""

from typing import Any

from pydantic import SecretStr, GetCoreSchemaHandler
from pydantic_core import CoreSchema, core_schema
from cryptography.fernet import Fernet


class FernetKey(SecretStr):
    """
    A Pydantic type for Fernet encryption keys.
    Extends SecretStr to ensure the key is never accidentally logged or exposed.
    Validates that the key is a valid Fernet key.
    """

    @classmethod
    def __get_pydantic_core_schema__(
        cls, source_type: Any, handler: GetCoreSchemaHandler
    ) -> CoreSchema:
        return core_schema.no_info_after_validator_function(
            cls._validate,
            core_schema.str_schema(),
            serialization=core_schema.plain_serializer_function_ser_schema(
                lambda x: "**********",
                info_arg=False,
                return_schema=core_schema.str_schema(),
            ),
        )

    @classmethod
    def _validate(cls, value: str) -> "FernetKey":
        """Validate that the value is a valid Fernet key."""
        if not value:
            raise ValueError("Fernet key cannot be empty")
        
        try:
            Fernet(value.encode() if isinstance(value, str) else value)
        except Exception as e:
            raise ValueError(f"Invalid Fernet key: {e}")
        
        instance = cls(value)
        return instance

    def get_fernet(self) -> Fernet:
        """Create a Fernet instance from this key."""
        key_bytes = self.get_secret_value().encode()
        return Fernet(key_bytes)

    @classmethod
    def generate(cls) -> "FernetKey":
        """Generate a new random Fernet key."""
        key = Fernet.generate_key().decode()
        return cls(key)
