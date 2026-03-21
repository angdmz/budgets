"""add_user_preferences

Revision ID: 61408a6af122
Revises: ddd261308f71
Create Date: 2026-03-21 19:30:21.399590

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '61408a6af122'
down_revision: Union[str, None] = 'ddd261308f71'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Create enum types
    theme_enum = sa.Enum('LIGHT', 'DIM', 'DARK', name='theme')
    theme_enum.create(op.get_bind(), checkfirst=True)

    language_enum = sa.Enum('EN', 'ES', name='language')
    language_enum.create(op.get_bind(), checkfirst=True)

    # currency enum already exists from initial schema

    op.create_table(
        'user_preferences',
        sa.Column('id', sa.Integer(), autoincrement=True, nullable=False),
        sa.Column('external_id', sa.dialects.postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column('user_id', sa.Integer(), sa.ForeignKey('users.id', ondelete='CASCADE'), nullable=False, unique=True),
        sa.Column('theme', theme_enum, nullable=False, server_default='LIGHT'),
        sa.Column('language', language_enum, nullable=False, server_default='EN'),
        sa.Column('display_currency', sa.Enum('USD', 'EUR', 'GBP', 'ARS', 'BRL', 'MXN', 'CLP', 'COP', 'PEN', 'UYU', name='currency', create_type=False), nullable=False, server_default='USD'),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
        sa.Column('updated_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
        sa.Column('revoked_at', sa.DateTime(timezone=True), nullable=True),
        sa.PrimaryKeyConstraint('id'),
    )
    op.create_index('ix_user_preferences_user', 'user_preferences', ['user_id'], unique=True)


def downgrade() -> None:
    op.drop_index('ix_user_preferences_user', table_name='user_preferences')
    op.drop_table('user_preferences')
    sa.Enum(name='language').drop(op.get_bind(), checkfirst=True)
    sa.Enum(name='theme').drop(op.get_bind(), checkfirst=True)
