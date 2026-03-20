"""
Encryptor utility for encrypting and decrypting data.
"""

from typing import Optional

from cryptography.fernet import Fernet

from .fernet_key import FernetKey


class Encryptor:
    """
    Utility class for encrypting and decrypting string data using Fernet.
    This implementation is compatible with fernet-go library.
    """

    def __init__(self, key: FernetKey):
        """
        Initialize the encryptor with a Fernet key.
        
        Args:
            key: A valid FernetKey instance
        """
        self._fernet = key.get_fernet()

    @classmethod
    def from_key_string(cls, key_string: str) -> "Encryptor":
        """
        Create an Encryptor from a key string.
        
        Args:
            key_string: A valid Fernet key string
            
        Returns:
            An Encryptor instance
        """
        return cls(FernetKey(key_string))

    def encrypt(self, plaintext: str) -> str:
        """
        Encrypt a plaintext string.
        
        Args:
            plaintext: The string to encrypt
            
        Returns:
            The encrypted string (base64 encoded)
        """
        if not plaintext:
            return ""
        
        plaintext_bytes = plaintext.encode("utf-8")
        encrypted_bytes = self._fernet.encrypt(plaintext_bytes)
        return encrypted_bytes.decode("utf-8")

    def decrypt(self, ciphertext: str) -> str:
        """
        Decrypt an encrypted string.
        
        Args:
            ciphertext: The encrypted string (base64 encoded)
            
        Returns:
            The decrypted plaintext string
        """
        if not ciphertext:
            return ""
        
        ciphertext_bytes = ciphertext.encode("utf-8")
        decrypted_bytes = self._fernet.decrypt(ciphertext_bytes)
        return decrypted_bytes.decode("utf-8")

    def encrypt_optional(self, plaintext: Optional[str]) -> Optional[str]:
        """
        Encrypt an optional plaintext string.
        
        Args:
            plaintext: The string to encrypt, or None
            
        Returns:
            The encrypted string, or None if input was None
        """
        if plaintext is None:
            return None
        return self.encrypt(plaintext)

    def decrypt_optional(self, ciphertext: Optional[str]) -> Optional[str]:
        """
        Decrypt an optional encrypted string.
        
        Args:
            ciphertext: The encrypted string, or None
            
        Returns:
            The decrypted string, or None if input was None
        """
        if ciphertext is None:
            return None
        return self.decrypt(ciphertext)
