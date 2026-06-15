"""make users email nullable

Users authenticated via Auth0 access tokens (username/password or social) may
not have an email claim in the access token payload. Previously the column was
NOT NULL with a unique index, which caused two separate problems:
  1. Empty strings "" were inserted when the token had no email claim.
  2. Subsequent requests from *different* Auth0 users (new test run, new sub)
     that also had no email claim would hit the unique constraint on email=""
     and receive a 500 from the API.

Fix: make the column nullable so that missing emails are stored as NULL.
PostgreSQL unique indexes treat each NULL as distinct, so multiple users
without an email never conflict.

Also converts any existing empty-string emails to NULL as part of the upgrade.

Revision ID: a3b4c5d6e7f8
Revises: 61408a6af122
Create Date: 2026-06-14 21:00:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'a3b4c5d6e7f8'
down_revision: Union[str, None] = '61408a6af122'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.alter_column('users', 'email',
                    existing_type=sa.String(length=255),
                    nullable=True)
    op.execute("UPDATE users SET email = NULL WHERE email = ''")


def downgrade() -> None:
    op.execute("UPDATE users SET email = '' WHERE email IS NULL")
    op.alter_column('users', 'email',
                    existing_type=sa.String(length=255),
                    nullable=False)
