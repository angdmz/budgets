package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/budgets/core/internal/domain"
)

type UserRepository interface {
	GetOrCreateByProvider(ctx context.Context, tx pgx.Tx, providerID string, provider domain.AuthProvider, email, displayName, avatarURL string) (*domain.User, error)
	GetByID(ctx context.Context, tx pgx.Tx, id int64) (*domain.User, error)
	GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.User, error)
}

type BudgetingGroupRepository interface {
	Create(ctx context.Context, tx pgx.Tx, group *domain.BudgetingGroup) error
	GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.BudgetingGroup, error)
	GetByParticipantUserID(ctx context.Context, tx pgx.Tx, userID int64) ([]domain.BudgetingGroup, error)
	Update(ctx context.Context, tx pgx.Tx, group *domain.BudgetingGroup) error
	Revoke(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) error
}

type ParticipantRepository interface {
	Create(ctx context.Context, tx pgx.Tx, participant *domain.Participant) error
	GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.Participant, error)
	GetByGroupID(ctx context.Context, tx pgx.Tx, groupID int64) ([]domain.Participant, error)
	Update(ctx context.Context, tx pgx.Tx, participant *domain.Participant) error
	Revoke(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) error
}

type UserParticipantRepository interface {
	Create(ctx context.Context, tx pgx.Tx, up *domain.UserParticipant) error
	GetByUserIDAndGroupID(ctx context.Context, tx pgx.Tx, userID int64, groupID int64) (*domain.UserParticipant, error)
	GetByUserIDAndParticipantID(ctx context.Context, tx pgx.Tx, userID, participantID int64) (*domain.UserParticipant, error)
}

type UserPreferenceRepository interface {
	GetByUserID(ctx context.Context, tx pgx.Tx, userID int64) (*domain.UserPreference, error)
	Upsert(ctx context.Context, tx pgx.Tx, pref *domain.UserPreference) error
}

type ExpenseCategoryRepository interface {
	Create(ctx context.Context, tx pgx.Tx, category *domain.ExpenseCategory) error
	GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.ExpenseCategory, error)
	GetByGroupID(ctx context.Context, tx pgx.Tx, groupID int64) ([]domain.ExpenseCategory, error)
	Update(ctx context.Context, tx pgx.Tx, category *domain.ExpenseCategory) error
	Revoke(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) error
}

type BudgetRepository interface {
	Create(ctx context.Context, tx pgx.Tx, budget *domain.Budget) error
	GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.Budget, error)
	GetByGroupID(ctx context.Context, tx pgx.Tx, groupID int64) ([]domain.Budget, error)
	Update(ctx context.Context, tx pgx.Tx, budget *domain.Budget) error
	Revoke(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) error
}

type ExpectedExpenseRepository interface {
	Create(ctx context.Context, tx pgx.Tx, expense *domain.ExpectedExpense, encryptedAmount string) error
	GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.ExpectedExpense, string, error)
	GetByBudgetID(ctx context.Context, tx pgx.Tx, budgetID int64) ([]domain.ExpectedExpense, []string, error)
	Update(ctx context.Context, tx pgx.Tx, expense *domain.ExpectedExpense, encryptedAmount string) error
	Revoke(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) error
}

type ActualExpenseRepository interface {
	Create(ctx context.Context, tx pgx.Tx, expense *domain.ActualExpense, encryptedAmount string) error
	GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.ActualExpense, string, error)
	GetByBudgetID(ctx context.Context, tx pgx.Tx, budgetID int64) ([]domain.ActualExpense, []string, error)
	Update(ctx context.Context, tx pgx.Tx, expense *domain.ActualExpense, encryptedAmount string) error
	Revoke(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) error
}
