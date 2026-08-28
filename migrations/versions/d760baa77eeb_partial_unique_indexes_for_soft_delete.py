"""partial unique indexes for soft delete

Revision ID: d760baa77eeb
Revises: 14ee733dea8f
Create Date: 2026-08-28 20:37:17.527945

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'd760baa77eeb'
down_revision: Union[str, None] = '14ee733dea8f'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Drop regular unique indexes and recreate as partial unique indexes
    # so soft-deleted rows (revoked_at IS NOT NULL) don't block re-creation.
    op.drop_index('ix_expense_categories_group_name', table_name='expense_categories')
    op.create_index(
        'ix_expense_categories_group_name',
        'expense_categories',
        ['budgeting_group_id', 'name'],
        unique=True,
        postgresql_where=sa.text('revoked_at IS NULL'),
    )
    op.drop_index('ix_participants_group_name', table_name='participants')
    op.create_index(
        'ix_participants_group_name',
        'participants',
        ['budgeting_group_id', 'name'],
        unique=True,
        postgresql_where=sa.text('revoked_at IS NULL'),
    )


def downgrade() -> None:
    op.drop_index('ix_participants_group_name', table_name='participants')
    op.create_index(
        'ix_participants_group_name',
        'participants',
        ['budgeting_group_id', 'name'],
        unique=True,
    )
    op.drop_index('ix_expense_categories_group_name', table_name='expense_categories')
    op.create_index(
        'ix_expense_categories_group_name',
        'expense_categories',
        ['budgeting_group_id', 'name'],
        unique=True,
    )