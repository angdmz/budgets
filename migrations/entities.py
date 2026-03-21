"""
SQLAlchemy models for the Budget Management System.
All models follow the conventions:
- Internal primary key: autoincrement integer named 'id'
- External identifier: UUID named 'external_id'
- Audit fields: created_at, updated_at, revoked_at
- Soft deletion via revoked_at
"""

from enum import Enum as PyEnum

from sqlalchemy import (
    Column,
    Integer,
    String,
    Text,
    TIMESTAMP,
    ForeignKey,
    Date,
    Index,
    Enum,
    func,
)
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import declarative_base, relationship

from encryption import Encryptor, EncryptedMoney
from secrets import get_secrets_provider, SecretNotFoundError

Base = declarative_base()
metadata = Base.metadata


class Currency(PyEnum):
    """Supported currencies."""
    USD = "USD"
    EUR = "EUR"
    GBP = "GBP"
    ARS = "ARS"
    BRL = "BRL"
    MXN = "MXN"
    CLP = "CLP"
    COP = "COP"
    PEN = "PEN"
    UYU = "UYU"


class AuthProvider(PyEnum):
    """Supported authentication providers."""
    GOOGLE = "google"
    GITHUB = "github"
    LOCAL = "local"


class Theme(PyEnum):
    """UI theme options."""
    LIGHT = "LIGHT"
    DIM = "DIM"
    DARK = "DARK"


class Language(PyEnum):
    """Supported languages."""
    EN = "EN"
    ES = "ES"


def _get_encryptor():
    """Get the encryptor instance, lazily initialized."""
    try:
        secrets_provider = get_secrets_provider()
        encryption_key = secrets_provider.get_secret("encryption_key")
    except SecretNotFoundError:
        import os
        encryption_key = os.environ.get("ENCRYPTION_KEY")
    
    if encryption_key:
        return Encryptor.from_key_string(encryption_key)
    return None


encryptor = _get_encryptor()


def get_encrypted_money_type():
    """Get the EncryptedMoney type with the configured encryptor."""
    if encryptor is None:
        raise RuntimeError(
            "Encryption key not configured. Set ENCRYPTION_KEY environment variable."
        )
    return EncryptedMoney(encryptor, length=1024)


class BaseModel(Base):
    """
    Abstract base model with audit fields.
    All models inherit from this to get consistent audit tracking.
    """
    __abstract__ = True

    created_at = Column(
        TIMESTAMP(timezone=True),
        nullable=False,
        server_default=func.current_timestamp(),
    )
    updated_at = Column(
        TIMESTAMP(timezone=True),
        nullable=False,
        server_default=func.current_timestamp(),
        onupdate=func.current_timestamp(),
    )
    revoked_at = Column(
        TIMESTAMP(timezone=True),
        nullable=True,
    )


class BaseModelWithID(BaseModel):
    """
    Abstract base model with internal ID and external UUID.
    Provides the standard primary key pattern for all entities.
    """
    __abstract__ = True

    id = Column(Integer, primary_key=True, autoincrement=True)
    external_id = Column(
        UUID(as_uuid=True),
        nullable=False,
        unique=True,
        index=True,
        server_default=func.gen_random_uuid(),
    )


class User(BaseModelWithID):
    """
    Represents an authenticated user in the system.
    Users can be members of multiple participants across different groups.
    """
    __tablename__ = "users"

    external_provider_id = Column(String(255), nullable=False, index=True)
    auth_provider = Column(Enum(AuthProvider), nullable=False, default=AuthProvider.GOOGLE)
    email = Column(String(255), nullable=False, unique=True, index=True)
    display_name = Column(String(255), nullable=True)
    avatar_url = Column(String(1024), nullable=True)

    user_participants = relationship("UserParticipant", back_populates="user")

    __table_args__ = (
        Index(
            "ix_users_provider_id",
            "auth_provider",
            "external_provider_id",
            unique=True,
        ),
    )


class BudgetingGroup(BaseModelWithID):
    """
    Main organizational unit for the budget system.
    All data belongs to a group, providing data isolation.
    """
    __tablename__ = "budgeting_groups"

    name = Column(String(255), nullable=False)
    description = Column(Text, nullable=True)

    participants = relationship("Participant", back_populates="budgeting_group")
    categories = relationship("ExpenseCategory", back_populates="budgeting_group")
    budgets = relationship("Budget", back_populates="budgeting_group")


class Participant(BaseModelWithID):
    """
    Represents a participant in a budgeting group.
    A participant is a business-level concept (e.g., a couple, a family unit).
    Multiple users can be associated with the same participant.
    """
    __tablename__ = "participants"

    name = Column(String(255), nullable=False)
    description = Column(Text, nullable=True)
    
    budgeting_group_id = Column(
        Integer,
        ForeignKey("budgeting_groups.id", ondelete="CASCADE"),
        nullable=False,
    )

    budgeting_group = relationship("BudgetingGroup", back_populates="participants")
    user_participants = relationship("UserParticipant", back_populates="participant")

    __table_args__ = (
        Index(
            "ix_participants_group_name",
            "budgeting_group_id",
            "name",
            unique=True,
        ),
    )


class UserParticipant(BaseModelWithID):
    """
    Association between Users and Participants.
    Allows multiple users to be part of the same participant (e.g., a couple).
    Contains additional attributes for the relationship.
    """
    __tablename__ = "user_participants"

    user_id = Column(
        Integer,
        ForeignKey("users.id", ondelete="CASCADE"),
        nullable=False,
    )
    participant_id = Column(
        Integer,
        ForeignKey("participants.id", ondelete="CASCADE"),
        nullable=False,
    )
    role = Column(String(50), nullable=False, default="member")
    is_primary = Column(Integer, nullable=False, default=0)

    user = relationship("User", back_populates="user_participants")
    participant = relationship("Participant", back_populates="user_participants")

    __table_args__ = (
        Index(
            "ix_user_participants_user_participant",
            "user_id",
            "participant_id",
            unique=True,
        ),
        Index("ix_user_participants_user", "user_id"),
        Index("ix_user_participants_participant", "participant_id"),
    )


class ExpenseCategory(BaseModelWithID):
    """
    Expense categories defined per budgeting group.
    Categories are not shared across groups.
    """
    __tablename__ = "expense_categories"

    name = Column(String(255), nullable=False)
    description = Column(Text, nullable=True)
    color = Column(String(7), nullable=True)
    icon = Column(String(50), nullable=True)

    budgeting_group_id = Column(
        Integer,
        ForeignKey("budgeting_groups.id", ondelete="CASCADE"),
        nullable=False,
    )

    budgeting_group = relationship("BudgetingGroup", back_populates="categories")
    expected_expenses = relationship("ExpectedExpense", back_populates="category")
    actual_expenses = relationship("ActualExpense", back_populates="category")

    __table_args__ = (
        Index(
            "ix_expense_categories_group_name",
            "budgeting_group_id",
            "name",
            unique=True,
        ),
    )


class Budget(BaseModelWithID):
    """
    Represents a budget period (typically a month or week).
    Contains expected and actual expenses.
    """
    __tablename__ = "budgets"

    name = Column(String(255), nullable=False)
    description = Column(Text, nullable=True)
    start_date = Column(Date, nullable=False)
    end_date = Column(Date, nullable=False)

    budgeting_group_id = Column(
        Integer,
        ForeignKey("budgeting_groups.id", ondelete="CASCADE"),
        nullable=False,
    )

    budgeting_group = relationship("BudgetingGroup", back_populates="budgets")
    expected_expenses = relationship("ExpectedExpense", back_populates="budget")
    actual_expenses = relationship("ActualExpense", back_populates="budget")

    __table_args__ = (
        Index("ix_budgets_group_dates", "budgeting_group_id", "start_date", "end_date"),
    )


class ExpectedExpense(BaseModelWithID):
    """
    Expected expense within a budget.
    Contains encrypted monetary value with currency.
    """
    __tablename__ = "expected_expenses"

    name = Column(String(255), nullable=False)
    description = Column(Text, nullable=True)
    
    encrypted_amount = Column(
        String(1024),
        nullable=False,
    )

    budget_id = Column(
        Integer,
        ForeignKey("budgets.id", ondelete="CASCADE"),
        nullable=False,
    )
    category_id = Column(
        Integer,
        ForeignKey("expense_categories.id", ondelete="SET NULL"),
        nullable=True,
    )

    budget = relationship("Budget", back_populates="expected_expenses")
    category = relationship("ExpenseCategory", back_populates="expected_expenses")
    actual_expenses = relationship("ActualExpense", back_populates="expected_expense")

    __table_args__ = (
        Index("ix_expected_expenses_budget", "budget_id"),
        Index("ix_expected_expenses_category", "category_id"),
    )


class ActualExpense(BaseModelWithID):
    """
    Actual expense recorded within a budget.
    Contains encrypted monetary value with currency and date.
    """
    __tablename__ = "actual_expenses"

    name = Column(String(255), nullable=False)
    description = Column(Text, nullable=True)
    expense_date = Column(Date, nullable=False)
    
    encrypted_amount = Column(
        String(1024),
        nullable=False,
    )

    budget_id = Column(
        Integer,
        ForeignKey("budgets.id", ondelete="CASCADE"),
        nullable=False,
    )
    category_id = Column(
        Integer,
        ForeignKey("expense_categories.id", ondelete="SET NULL"),
        nullable=True,
    )
    expected_expense_id = Column(
        Integer,
        ForeignKey("expected_expenses.id", ondelete="SET NULL"),
        nullable=True,
    )

    budget = relationship("Budget", back_populates="actual_expenses")
    category = relationship("ExpenseCategory", back_populates="actual_expenses")
    expected_expense = relationship("ExpectedExpense", back_populates="actual_expenses")

    __table_args__ = (
        Index("ix_actual_expenses_budget", "budget_id"),
        Index("ix_actual_expenses_category", "category_id"),
        Index("ix_actual_expenses_expected", "expected_expense_id"),
        Index("ix_actual_expenses_date", "expense_date"),
    )


class UserPreference(BaseModelWithID):
    """
    Stores per-user preferences such as theme, language, and display currency.
    One row per user (unique constraint on user_id).
    """
    __tablename__ = "user_preferences"

    user_id = Column(
        Integer,
        ForeignKey("users.id", ondelete="CASCADE"),
        nullable=False,
        unique=True,
    )
    theme = Column(
        Enum(Theme),
        nullable=False,
        default=Theme.LIGHT,
        server_default="LIGHT",
    )
    language = Column(
        Enum(Language),
        nullable=False,
        default=Language.EN,
        server_default="EN",
    )
    display_currency = Column(
        Enum(Currency),
        nullable=False,
        default=Currency.USD,
        server_default="USD",
    )

    user = relationship("User")

    __table_args__ = (
        Index("ix_user_preferences_user", "user_id", unique=True),
    )
