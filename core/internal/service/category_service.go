package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/repository"
)

type CategoryService interface {
	Create(ctx context.Context, groupExternalID uuid.UUID, name, description, color, icon string, userID int64) (*domain.ExpenseCategory, error)
	GetByGroupID(ctx context.Context, groupExternalID uuid.UUID, userID int64) ([]domain.ExpenseCategory, error)
	Update(ctx context.Context, externalID uuid.UUID, name, description, color, icon string, userID int64) (*domain.ExpenseCategory, error)
	Delete(ctx context.Context, externalID uuid.UUID, userID int64) error
}

type categoryService struct {
	pool                *pgxpool.Pool
	categoryRepo        repository.ExpenseCategoryRepository
	groupRepo           repository.BudgetingGroupRepository
	userParticipantRepo repository.UserParticipantRepository
}

func NewCategoryService(
	pool *pgxpool.Pool,
	categoryRepo repository.ExpenseCategoryRepository,
	groupRepo repository.BudgetingGroupRepository,
	userParticipantRepo repository.UserParticipantRepository,
) CategoryService {
	return &categoryService{
		pool:                pool,
		categoryRepo:        categoryRepo,
		groupRepo:           groupRepo,
		userParticipantRepo: userParticipantRepo,
	}
}

func (s *categoryService) Create(ctx context.Context, groupExternalID uuid.UUID, name, description, color, icon string, userID int64) (*domain.ExpenseCategory, error) {
	var category *domain.ExpenseCategory

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		group, err := s.groupRepo.GetByExternalID(ctx, tx, groupExternalID)
		if err != nil {
			return err
		}

		_, err = s.userParticipantRepo.GetByUserIDAndGroupID(ctx, tx, userID, group.ID)
		if err != nil {
			return domain.ErrForbidden
		}

		category = &domain.ExpenseCategory{
			Name:             name,
			Description:      description,
			Color:            color,
			Icon:             icon,
			BudgetingGroupID: group.ID,
		}

		return s.categoryRepo.Create(ctx, tx, category)
	})

	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *categoryService) GetByGroupID(ctx context.Context, groupExternalID uuid.UUID, userID int64) ([]domain.ExpenseCategory, error) {
	var categories []domain.ExpenseCategory

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		group, err := s.groupRepo.GetByExternalID(ctx, tx, groupExternalID)
		if err != nil {
			return err
		}

		_, err = s.userParticipantRepo.GetByUserIDAndGroupID(ctx, tx, userID, group.ID)
		if err != nil {
			return domain.ErrForbidden
		}

		categories, err = s.categoryRepo.GetByGroupID(ctx, tx, group.ID)
		return err
	})

	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *categoryService) Update(ctx context.Context, externalID uuid.UUID, name, description, color, icon string, userID int64) (*domain.ExpenseCategory, error) {
	var category *domain.ExpenseCategory

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		category, err = s.categoryRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		_, err = s.userParticipantRepo.GetByUserIDAndGroupID(ctx, tx, userID, category.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		category.Name = name
		category.Description = description
		category.Color = color
		category.Icon = icon

		return s.categoryRepo.Update(ctx, tx, category)
	})

	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *categoryService) Delete(ctx context.Context, externalID uuid.UUID, userID int64) error {
	return database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		category, err := s.categoryRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		_, err = s.userParticipantRepo.GetByUserIDAndGroupID(ctx, tx, userID, category.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		return s.categoryRepo.Revoke(ctx, tx, externalID)
	})
}
