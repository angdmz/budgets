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
    # Create enum types via raw SQL with IF NOT EXISTS safety
    conn = op.get_bind()
    conn.execute(sa.text("DO $$ BEGIN CREATE TYPE theme AS ENUM ('LIGHT', 'DIM', 'DARK'); EXCEPTION WHEN duplicate_object THEN null; END $$;"))
    conn.execute(sa.text("DO $$ BEGIN CREATE TYPE language AS ENUM ('EN', 'ES'); EXCEPTION WHEN duplicate_object THEN null; END $$;"))
    conn.execute(sa.text("DO $$ BEGIN CREATE TYPE currency AS ENUM ('USD', 'EUR', 'GBP', 'ARS', 'BRL', 'MXN', 'CLP', 'COP', 'PEN', 'UYU'); EXCEPTION WHEN duplicate_object THEN null; END $$;"))

    # Create table using raw column types to avoid SQLAlchemy re-creating enums
    op.execute("""
        CREATE TABLE user_preferences (
            id SERIAL PRIMARY KEY,
            external_id UUID NOT NULL DEFAULT gen_random_uuid(),
            user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
            theme theme NOT NULL DEFAULT 'LIGHT',
            language language NOT NULL DEFAULT 'EN',
            display_currency currency NOT NULL DEFAULT 'USD',
            created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
            revoked_at TIMESTAMPTZ
        );
    """)
    op.create_index('ix_user_preferences_user', 'user_preferences', ['user_id'], unique=True)


def downgrade() -> None:
    op.drop_index('ix_user_preferences_user', table_name='user_preferences')
    op.drop_table('user_preferences')
    sa.Enum(name='currency').drop(op.get_bind(), checkfirst=True)
    sa.Enum(name='language').drop(op.get_bind(), checkfirst=True)
    sa.Enum(name='theme').drop(op.get_bind(), checkfirst=True)
