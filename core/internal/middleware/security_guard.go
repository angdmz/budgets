package middleware

import (
	"context"

	"github.com/google/uuid"

	"github.com/budgets/core/internal/domain"
)

// SecurityGuard performs authorization checks against persisted state.
// Authorization is an infrastructure concern: it lives here, not on domain
// objects. The queries run inside the caller's transaction via Persister.
type SecurityGuard interface {
	AuthorizeGroupAccess(ctx context.Context, p domain.Persister, groupExternalID uuid.UUID) error
	AuthorizeGroupOwnership(ctx context.Context, p domain.Persister, groupExternalID uuid.UUID) error
	AuthorizeInvitationOwnership(ctx context.Context, p domain.Persister, invitationExternalID uuid.UUID) error
	AuthorizeBudgetAccess(ctx context.Context, p domain.Persister, budgetExternalID uuid.UUID) error
	AuthorizeCategoryAccess(ctx context.Context, p domain.Persister, categoryExternalID uuid.UUID) error
	AuthorizeExpenseAccess(ctx context.Context, p domain.Persister, expenseExternalID uuid.UUID) error
}

type securityGuard struct {
	user *domain.PersistedUser
}

func NewSecurityGuard(user *domain.PersistedUser) SecurityGuard {
	return &securityGuard{user: user}
}

func (s *securityGuard) AuthorizeGroupAccess(ctx context.Context, p domain.Persister, groupExternalID uuid.UUID) error {
	var exists bool
	err := p.QueryRow(
		ctx,
		[]any{&exists},
		`SELECT EXISTS(
			SELECT 1 FROM user_participants up
			JOIN participants pt ON up.participant_id = pt.id
			JOIN budgeting_groups bg ON pt.budgeting_group_id = bg.id
			WHERE up.user_id = $1 AND bg.external_id = $2
			AND up.revoked_at IS NULL AND pt.revoked_at IS NULL
		)`,
		s.user, groupExternalID,
	)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrForbidden
	}
	return nil
}

func (s *securityGuard) AuthorizeGroupOwnership(ctx context.Context, p domain.Persister, groupExternalID uuid.UUID) error {
	var isOwner bool
	err := p.QueryRow(
		ctx,
		[]any{&isOwner},
		`SELECT EXISTS(
			SELECT 1 FROM user_participants up
			JOIN participants pt ON up.participant_id = pt.id
			JOIN budgeting_groups bg ON pt.budgeting_group_id = bg.id
			WHERE up.user_id = $1 AND bg.external_id = $2 AND up.role = 'owner'
			AND up.revoked_at IS NULL AND pt.revoked_at IS NULL AND bg.revoked_at IS NULL
		)`,
		s.user, groupExternalID,
	)
	if err != nil {
		return err
	}
	if !isOwner {
		return domain.ErrForbidden
	}
	return nil
}

func (s *securityGuard) AuthorizeBudgetAccess(ctx context.Context, p domain.Persister, budgetExternalID uuid.UUID) error {
	var budgetExists bool
	err := p.QueryRow(
		ctx,
		[]any{&budgetExists},
		`SELECT EXISTS(SELECT 1 FROM budgets WHERE external_id = $1 AND revoked_at IS NULL)`,
		budgetExternalID,
	)
	if err != nil {
		return err
	}
	if !budgetExists {
		return domain.ErrNotFound
	}

	var hasAccess bool
	err = p.QueryRow(
		ctx,
		[]any{&hasAccess},
		`SELECT EXISTS(
			SELECT 1 FROM user_participants up
			JOIN participants pt ON up.participant_id = pt.id
			JOIN budgets b ON b.budgeting_group_id = pt.budgeting_group_id
			WHERE up.user_id = $1 AND b.external_id = $2
			AND up.revoked_at IS NULL AND pt.revoked_at IS NULL AND b.revoked_at IS NULL
		)`,
		s.user, budgetExternalID,
	)
	if err != nil {
		return err
	}
	if !hasAccess {
		return domain.ErrForbidden
	}
	return nil
}

func (s *securityGuard) AuthorizeInvitationOwnership(ctx context.Context, p domain.Persister, invitationExternalID uuid.UUID) error {
	var isOwner bool
	err := p.QueryRow(
		ctx,
		[]any{&isOwner},
		`SELECT EXISTS(
			SELECT 1 FROM user_participants up
			JOIN participants pt ON up.participant_id = pt.id
			JOIN group_invitations gi ON gi.budgeting_group_id = pt.budgeting_group_id
			WHERE up.user_id = $1 AND gi.external_id = $2 AND up.role = 'owner'
			AND up.revoked_at IS NULL AND pt.revoked_at IS NULL AND gi.revoked_at IS NULL
		)`,
		s.user, invitationExternalID,
	)
	if err != nil {
		return err
	}
	if !isOwner {
		return domain.ErrForbidden
	}
	return nil
}

func (s *securityGuard) AuthorizeCategoryAccess(ctx context.Context, p domain.Persister, categoryExternalID uuid.UUID) error {
	var categoryExists bool
	err := p.QueryRow(
		ctx,
		[]any{&categoryExists},
		`SELECT EXISTS(SELECT 1 FROM expense_categories WHERE external_id = $1 AND revoked_at IS NULL)`,
		categoryExternalID,
	)
	if err != nil {
		return err
	}
	if !categoryExists {
		return domain.ErrNotFound
	}

	var hasAccess bool
	err = p.QueryRow(
		ctx,
		[]any{&hasAccess},
		`SELECT EXISTS(
			SELECT 1 FROM user_participants up
			JOIN participants pt ON up.participant_id = pt.id
			JOIN expense_categories ec ON ec.budgeting_group_id = pt.budgeting_group_id
			WHERE up.user_id = $1 AND ec.external_id = $2
			AND up.revoked_at IS NULL AND pt.revoked_at IS NULL AND ec.revoked_at IS NULL
		)`,
		s.user, categoryExternalID,
	)
	if err != nil {
		return err
	}
	if !hasAccess {
		return domain.ErrForbidden
	}
	return nil
}

func (s *securityGuard) AuthorizeExpenseAccess(ctx context.Context, p domain.Persister, expenseExternalID uuid.UUID) error {
	var expenseExists bool
	err := p.QueryRow(
		ctx,
		[]any{&expenseExists},
		`SELECT EXISTS(
			SELECT 1 FROM expected_expenses WHERE external_id = $1 AND revoked_at IS NULL
			UNION
			SELECT 1 FROM actual_expenses WHERE external_id = $1 AND revoked_at IS NULL
		)`,
		expenseExternalID,
	)
	if err != nil {
		return err
	}
	if !expenseExists {
		return domain.ErrNotFound
	}

	var hasAccess bool
	err = p.QueryRow(
		ctx,
		[]any{&hasAccess},
		`SELECT EXISTS(
			SELECT 1 FROM user_participants up
			JOIN participants pt ON up.participant_id = pt.id
			LEFT JOIN expected_expenses ee ON ee.budget_id IN (
				SELECT id FROM budgets WHERE budgeting_group_id = pt.budgeting_group_id AND revoked_at IS NULL
			) AND ee.external_id = $2 AND ee.revoked_at IS NULL
			LEFT JOIN actual_expenses ae ON ae.budget_id IN (
				SELECT id FROM budgets WHERE budgeting_group_id = pt.budgeting_group_id AND revoked_at IS NULL
			) AND ae.external_id = $2 AND ae.revoked_at IS NULL
			WHERE up.user_id = $1 AND (ee.id IS NOT NULL OR ae.id IS NOT NULL)
			AND up.revoked_at IS NULL AND pt.revoked_at IS NULL
		)`,
		s.user, expenseExternalID,
	)
	if err != nil {
		return err
	}
	if !hasAccess {
		return domain.ErrForbidden
	}
	return nil
}
