"""
SQLAlchemy composite type for encrypted monetary values with currency.
"""

import json
from decimal import Decimal
from typing import Optional, Any, Tuple

from sqlalchemy import String, TypeDecorator

from .encryptor import Encryptor


class EncryptedMoney(TypeDecorator):
    """
    SQLAlchemy type that stores encrypted monetary values with currency.
    
    Money is stored as a JSON object containing:
    - amount: The monetary amount as a string (to preserve precision)
    - currency: The ISO 4217 currency code
    
    The entire JSON is encrypted before storage.
    
    Usage:
        class Expense(Base):
            amount = Column(EncryptedMoney(encryptor), nullable=False)
    """

    impl = String
    cache_ok = False

    def __init__(self, encryptor: Encryptor, length: Optional[int] = None):
        """
        Initialize the encrypted money type.
        
        Args:
            encryptor: An Encryptor instance for encryption/decryption
            length: Optional maximum length for the underlying String column
        """
        self._encryptor = encryptor
        super().__init__(length=length)

    def process_bind_param(
        self, value: Optional[Tuple[Decimal, str]], dialect: Any
    ) -> Optional[str]:
        """
        Encrypt the monetary value before storing in the database.
        
        Args:
            value: A tuple of (amount, currency_code) or None
            dialect: The SQLAlchemy dialect
            
        Returns:
            The encrypted JSON string
        """
        if value is None:
            return None
        
        amount, currency = value
        money_data = {
            "amount": str(amount),
            "currency": currency.upper(),
        }
        json_str = json.dumps(money_data)
        return self._encryptor.encrypt(json_str)

    def process_result_value(
        self, value: Optional[str], dialect: Any
    ) -> Optional[Tuple[Decimal, str]]:
        """
        Decrypt the monetary value when retrieving from the database.
        
        Args:
            value: The encrypted JSON string from the database
            dialect: The SQLAlchemy dialect
            
        Returns:
            A tuple of (amount, currency_code) or None
        """
        if value is None:
            return None
        
        json_str = self._encryptor.decrypt(value)
        money_data = json.loads(json_str)
        
        amount = Decimal(money_data["amount"])
        currency = money_data["currency"]
        
        return (amount, currency)

    def copy(self, **kwargs: Any) -> "EncryptedMoney":
        """Create a copy of this type."""
        return EncryptedMoney(self._encryptor, self.impl.length)


class Money:
    """
    Value object representing a monetary amount with currency.
    This is a helper class for working with EncryptedMoney columns.
    """

    def __init__(self, amount: Decimal, currency: str):
        """
        Initialize a Money value.
        
        Args:
            amount: The monetary amount
            currency: The ISO 4217 currency code (e.g., "USD", "EUR")
        """
        if not isinstance(amount, Decimal):
            amount = Decimal(str(amount))
        
        self._amount = amount
        self._currency = currency.upper()

    @property
    def amount(self) -> Decimal:
        """Get the monetary amount."""
        return self._amount

    @property
    def currency(self) -> str:
        """Get the currency code."""
        return self._currency

    def to_tuple(self) -> Tuple[Decimal, str]:
        """Convert to a tuple for storage."""
        return (self._amount, self._currency)

    @classmethod
    def from_tuple(cls, value: Tuple[Decimal, str]) -> "Money":
        """Create a Money instance from a tuple."""
        return cls(amount=value[0], currency=value[1])

    def __eq__(self, other: Any) -> bool:
        if not isinstance(other, Money):
            return False
        return self._amount == other._amount and self._currency == other._currency

    def __repr__(self) -> str:
        return f"Money({self._amount}, '{self._currency}')"

    def __str__(self) -> str:
        return f"{self._amount} {self._currency}"
