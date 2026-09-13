"""add user onboarding tables

Revision ID: b1c2d3e4f5a6
Revises: 7ff1be9357c4
Create Date: 2026-09-10 02:00:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects.postgresql import UUID, JSONB


# revision identifiers, used by Alembic.
revision: str = 'b1c2d3e4f5a6'
down_revision: Union[str, None] = '7ff1be9357c4'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Create ENUM types
    op.execute("""
        DO $$ BEGIN
            CREATE TYPE onboardingstatus AS ENUM ('in_progress', 'completed', 'skipped');
        EXCEPTION WHEN duplicate_object THEN NULL; END $$;
    """)
    op.execute("""
        DO $$ BEGIN
            CREATE TYPE onboardingstep AS ENUM (
                'welcome',
                'choose_group',
                'choose_cadence',
                'add_expected_expenses',
                'budget_summary',
                'register_actual_expense',
                'compare_expenses',
                'dashboard_tour',
                'complete'
            );
        EXCEPTION WHEN duplicate_object THEN NULL; END $$;
    """)
    op.execute("""
        DO $$ BEGIN
            CREATE TYPE onboardingstepstatus AS ENUM ('pending', 'completed', 'skipped');
        EXCEPTION WHEN duplicate_object THEN NULL; END $$;
    """)

    # Create user_onboardings table
    op.execute("""
        CREATE TABLE IF NOT EXISTS user_onboardings (
            id SERIAL PRIMARY KEY,
            external_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
            user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            status onboardingstatus NOT NULL DEFAULT 'in_progress',
            current_step onboardingstep NOT NULL DEFAULT 'welcome',
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            revoked_at TIMESTAMP WITH TIME ZONE
        )
    """)
    op.execute("""
        CREATE UNIQUE INDEX IF NOT EXISTS ix_user_onboardings_user
        ON user_onboardings (user_id)
        WHERE revoked_at IS NULL
    """)

    # Create user_onboarding_steps table
    op.execute("""
        CREATE TABLE IF NOT EXISTS user_onboarding_steps (
            id SERIAL PRIMARY KEY,
            external_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
            user_onboarding_id INTEGER NOT NULL REFERENCES user_onboardings(id) ON DELETE CASCADE,
            step onboardingstep NOT NULL,
            status onboardingstepstatus NOT NULL DEFAULT 'pending',
            data JSONB,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            revoked_at TIMESTAMP WITH TIME ZONE
        )
    """)
    op.execute("""
        CREATE UNIQUE INDEX IF NOT EXISTS ix_user_onboarding_steps_onboarding_step
        ON user_onboarding_steps (user_onboarding_id, step)
        WHERE revoked_at IS NULL
    """)
    op.execute("""
        CREATE INDEX IF NOT EXISTS ix_user_onboarding_steps_onboarding
        ON user_onboarding_steps (user_onboarding_id)
    """)


def downgrade() -> None:
    op.execute("DROP TABLE IF EXISTS user_onboarding_steps")
    op.execute("DROP TABLE IF EXISTS user_onboardings")
    op.execute("DROP TYPE IF EXISTS onboardingstepstatus")
    op.execute("DROP TYPE IF EXISTS onboardingstep")
    op.execute("DROP TYPE IF EXISTS onboardingstatus")
