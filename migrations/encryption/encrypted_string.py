"""
SQLAlchemy custom type for encrypted strings.
"""

from typing import Optional, Any

from sqlalchemy import String, TypeDecorator

from .encryptor import Encryptor


class EncryptedString(TypeDecorator):
    """
    SQLAlchemy type that transparently encrypts/decrypts string values.
    
    Values are encrypted when stored in the database and decrypted when retrieved.
    Uses Fernet encryption which is compatible with both Python and Go.
    
    Usage:
        class MyModel(Base):
            sensitive_data = Column(EncryptedString(encryptor), nullable=False)
    """

    impl = String
    cache_ok = False

    def __init__(self, encryptor: Encryptor, length: Optional[int] = None):
        """
        Initialize the encrypted string type.
        
        Args:
            encryptor: An Encryptor instance for encryption/decryption
            length: Optional maximum length for the underlying String column
        """
        self._encryptor = encryptor
        super().__init__(length=length)

    def process_bind_param(self, value: Optional[str], dialect: Any) -> Optional[str]:
        """
        Encrypt the value before storing in the database.
        
        Args:
            value: The plaintext value to encrypt
            dialect: The SQLAlchemy dialect
            
        Returns:
            The encrypted value
        """
        if value is None:
            return None
        return self._encryptor.encrypt(value)

    def process_result_value(self, value: Optional[str], dialect: Any) -> Optional[str]:
        """
        Decrypt the value when retrieving from the database.
        
        Args:
            value: The encrypted value from the database
            dialect: The SQLAlchemy dialect
            
        Returns:
            The decrypted plaintext value
        """
        if value is None:
            return None
        return self._encryptor.decrypt(value)

    def copy(self, **kwargs: Any) -> "EncryptedString":
        """Create a copy of this type."""
        return EncryptedString(self._encryptor, self.impl.length)
