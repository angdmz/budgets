package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/repository"
)

type BudgetService interface {
	Create(ctx context.Context, groupExternalID uuid.UUID, name, description string, startDate, endDate time.Time, userID string) (*domain.Budget, error)
	GetByID(ctx context.Context, externalID uuid.UUID, userID string) (*domain.Budget, error)
	GetByGroupID(ctx context.Context, groupExternalID uuid.UUID, userID string) ([]domain.Budget, error)
	Update(ctx context.Context, externalID uuid.UUID, name, description string, startDate, endDate time.Time, userID string) (*domain.Budget, error)
	Delete(ctx context.Context, externalID uuid.UUID, userID string) error
}

type budgetService struct {
	pool            *pgxpool.Pool
	budgetRepo      repository.BudgetRepository
	groupRepo       repository.BudgetingGroupRepository
	participantRepo repository.ParticipantRepository
}

func NewBudgetService(
	pool *pgxpool.Pool,
	budgetRepo repository.BudgetRepository,
	groupRepo repository.BudgetingGroupRepository,
	participantRepo repository.ParticipantRepository,
) BudgetService {
	return &budgetService{
		pool:            pool,
		budgetRepo:      budgetRepo,
		groupRepo:       groupRepo,
		participantRepo: participantRepo,
	}
}

func (s *budgetService) Create(ctx context.Context, groupExternalID uuid.UUID, name, description string, startDate, endDate time.Time, userID string) (*domain.Budget, error) {
	var budget *domain.Budget

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		group, err := s.groupRepo.GetByExternalID(ctx, tx, groupExternalID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, group.ID)
		if err != nil {
			return domain.ErrForbidden
		}

		budget = &domain.Budget{
			Name:             name,
			Description:      description,
			StartDate:        startDate,
			EndDate:          endDate,
			BudgetingGroupID: group.ID,
		}

		return s.budgetRepo.Create(ctx, tx, budget)
	})

	if err != nil {
		return nil, err
	}

	return budget, nil
}

func (s *budgetService) GetByID(ctx context.Context, externalID uuid.UUID, userID string) (*domain.Budget, error) {
	var budget *domain.Budget

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		budget, err = s.budgetRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, budget.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return budget, nil
}

func (s *budgetService) GetByGroupID(ctx context.Context, groupExternalID uuid.UUID, userID string) ([]domain.Budget, error) {
	var budgets []domain.Budget

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		group, err := s.groupRepo.GetByExternalID(ctx, tx, groupExternalID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, group.ID)
		if err != nil {
			return domain.ErrForbidden
		}

		budgets, err = s.budgetRepo.GetByGroupID(ctx, tx, group.ID)
		return err
	})

	if err != nil {
		return nil, err
	}

	return budgets, nil
}

func (s *budgetService) Update(ctx context.Context, externalID uuid.UUID, name, description string, startDate, endDate time.Time, userID string) (*domain.Budget, error) {
	var budget *domain.Budget

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		budget, err = s.budgetRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, budget.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		budget.Name = name
		budget.Description = description
		budget.StartDate = startDate
		budget.EndDate = endDate

		return s.budgetRepo.Update(ctx, tx, budget)
	})

	if err != nil {
		return nil, err
	}

	return budget, nil
}

func (s *budgetService) Delete(ctx context.Context, externalID uuid.UUID, userID string) error {
	return database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		budget, err := s.budgetRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, budget.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		return s.budgetRepo.Revoke(ctx, tx, externalID)
	})
}
