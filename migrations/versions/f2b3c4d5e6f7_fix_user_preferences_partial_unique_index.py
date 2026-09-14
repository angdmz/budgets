"""fix user_preferences partial unique index

Revision ID: f2b3c4d5e6f7
Revises: e1a2b3c4d5e6
Create Date: 2026-09-14 15:10:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'f2b3c4d5e6f7'
down_revision: Union[str, None] = 'e1a2b3c4d5e6'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # The ON CONFLICT (user_id) WHERE revoked_at IS NULL clause in the Go
    # upsert requires a *partial* unique index. The original migration
    # (61408a6af122) created a plain unique index + column-level UNIQUE,
    # and d760baa77eeb missed user_preferences when adding partial indexes
    # to other soft-deletable tables.

    # Drop the plain unique index
    op.drop_index('ix_user_preferences_user', table_name='user_preferences')

    # Drop the column-level UNIQUE constraint on user_id
    op.execute('ALTER TABLE user_preferences DROP CONSTRAINT IF EXISTS user_preferences_user_id_key')

    # Recreate as a partial unique index
    op.create_index(
        'ix_user_preferences_user',
        'user_preferences',
        ['user_id'],
        unique=True,
        postgresql_where=sa.text('revoked_at IS NULL'),
    )


def downgrade() -> None:
    op.drop_index('ix_user_preferences_user', table_name='user_preferences')
    op.execute('ALTER TABLE user_preferences ADD CONSTRAINT user_preferences_user_id_key UNIQUE (user_id)')
    op.create_index(
        'ix_user_preferences_user',
        'user_preferences',
        ['user_id'],
        unique=True,
    )
