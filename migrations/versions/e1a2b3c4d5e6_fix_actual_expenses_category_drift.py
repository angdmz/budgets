"""fix actual_expenses category_id drift

Revision ID: e1a2b3c4d5e6
Revises: b1c2d3e4f5a6
Create Date: 2026-09-14 15:00:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'e1a2b3c4d5e6'
down_revision: Union[str, None] = 'b1c2d3e4f5a6'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # The money migration (7ff1be9357c4) silently reverted the NOT NULL + CASCADE
    # constraints that 14ee733dea8f had applied. Re-apply them here.
    #
    # Before running this, ensure no NULL category_id rows exist:
    #   SELECT count(*) FROM actual_expenses WHERE category_id IS NULL;
    # If any are found, backfill or delete them first.

    op.drop_constraint('actual_expenses_category_id_fkey', 'actual_expenses', type_='foreignkey')
    op.create_foreign_key(
        'actual_expenses_category_id_fkey',
        'actual_expenses',
        'expense_categories',
        ['category_id'],
        ['id'],
        ondelete='CASCADE',
    )
    op.alter_column(
        'actual_expenses',
        'category_id',
        existing_type=sa.INTEGER(),
        nullable=False,
    )


def downgrade() -> None:
    op.alter_column(
        'actual_expenses',
        'category_id',
        existing_type=sa.INTEGER(),
        nullable=True,
    )
    op.drop_constraint('actual_expenses_category_id_fkey', 'actual_expenses', type_='foreignkey')
    op.create_foreign_key(
        None,
        'actual_expenses',
        'expense_categories',
        ['category_id'],
        ['id'],
        ondelete='SET NULL',
    )
