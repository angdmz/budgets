"""add exchange rates and quote types

Revision ID: a3c4d5e6f7a8
Revises: f2b3c4d5e6f7
Create Date: 2026-09-14 15:20:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'a3c4d5e6f7a8'
down_revision: Union[str, None] = 'f2b3c4d5e6f7'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Create quote_type enum (uppercase labels, matching Currency/Theme/Language convention)
    op.execute("""
        DO $$ BEGIN
            CREATE TYPE quote_type AS ENUM ('OFFICIAL', 'BLUE', 'MEP', 'CCL', 'CRYPTO');
        EXCEPTION WHEN duplicate_object THEN NULL; END $$;
    """)

    # Create exchange_rates table
    op.execute("""
        CREATE TABLE IF NOT EXISTS exchange_rates (
            id SERIAL PRIMARY KEY,
            external_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
            from_currency currency NOT NULL,
            to_currency currency NOT NULL,
            quote quote_type NOT NULL,
            rate NUMERIC(20, 10) NOT NULL,
            provider VARCHAR(50) NOT NULL,
            observed_at TIMESTAMP WITH TIME ZONE NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            revoked_at TIMESTAMP WITH TIME ZONE
        )
    """)

    # Partial unique index on (from_currency, to_currency, quote) where not revoked
    op.execute("""
        CREATE UNIQUE INDEX IF NOT EXISTS ix_exchange_rates_pair_quote
        ON exchange_rates (from_currency, to_currency, quote)
        WHERE revoked_at IS NULL
    """)

    # Index on observed_at for rate history queries
    op.execute("""
        CREATE INDEX IF NOT EXISTS ix_exchange_rates_observed
        ON exchange_rates (observed_at)
    """)

    # Add preferred_quote_type column to user_preferences
    op.execute("""
        ALTER TABLE user_preferences
        ADD COLUMN IF NOT EXISTS preferred_quote_type quote_type NOT NULL DEFAULT 'OFFICIAL'
    """)


def downgrade() -> None:
    op.execute('ALTER TABLE user_preferences DROP COLUMN IF EXISTS preferred_quote_type')
    op.drop_index('ix_exchange_rates_observed', table_name='exchange_rates')
    op.drop_index('ix_exchange_rates_pair_quote', table_name='exchange_rates')
    op.execute('DROP TABLE IF EXISTS exchange_rates')
    op.execute('DROP TYPE IF EXISTS quote_type')
