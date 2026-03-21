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

type GroupService interface {
	Create(ctx context.Context, name, description string, userID int64, displayName string) (*domain.BudgetingGroup, error)
	GetByID(ctx context.Context, externalID uuid.UUID, userID int64) (*domain.BudgetingGroup, error)
	GetByUserID(ctx context.Context, userID int64) ([]domain.BudgetingGroup, error)
	Update(ctx context.Context, externalID uuid.UUID, name, description string, userID int64) (*domain.BudgetingGroup, error)
	Delete(ctx context.Context, externalID uuid.UUID, userID int64) error
}

type groupService struct {
	pool                *pgxpool.Pool
	groupRepo           repository.BudgetingGroupRepository
	participantRepo     repository.ParticipantRepository
	userParticipantRepo repository.UserParticipantRepository
}

func NewGroupService(
	pool *pgxpool.Pool,
	groupRepo repository.BudgetingGroupRepository,
	participantRepo repository.ParticipantRepository,
	userParticipantRepo repository.UserParticipantRepository,
) GroupService {
	return &groupService{
		pool:                pool,
		groupRepo:           groupRepo,
		participantRepo:     participantRepo,
		userParticipantRepo: userParticipantRepo,
	}
}

func (s *groupService) Create(ctx context.Context, name, description string, userID int64, displayName string) (*domain.BudgetingGroup, error) {
	var group *domain.BudgetingGroup

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		group = &domain.BudgetingGroup{
			Name:        name,
			Description: description,
		}

		if err := s.groupRepo.Create(ctx, tx, group); err != nil {
			return err
		}

		// Create an individual participant for the user in this group
		participant := &domain.Participant{
			Name:             displayName,
			BudgetingGroupID: group.ID,
		}

		if err := s.participantRepo.Create(ctx, tx, participant); err != nil {
			return err
		}

		// Link the user to the participant
		up := &domain.UserParticipant{
			UserID:        userID,
			ParticipantID: participant.ID,
			Role:          "owner",
			IsPrimary:     true,
		}

		return s.userParticipantRepo.Create(ctx, tx, up)
	})

	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *groupService) GetByID(ctx context.Context, externalID uuid.UUID, userID int64) (*domain.BudgetingGroup, error) {
	var group *domain.BudgetingGroup

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		group, err = s.groupRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		_, err = s.userParticipantRepo.GetByUserIDAndGroupID(ctx, tx, userID, group.ID)
		if err != nil {
			return domain.ErrForbidden
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *groupService) GetByUserID(ctx context.Context, userID int64) ([]domain.BudgetingGroup, error) {
	var groups []domain.BudgetingGroup

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		groups, err = s.groupRepo.GetByParticipantUserID(ctx, tx, userID)
		return err
	})

	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (s *groupService) Update(ctx context.Context, externalID uuid.UUID, name, description string, userID int64) (*domain.BudgetingGroup, error) {
	var group *domain.BudgetingGroup

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		group, err = s.groupRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		_, err = s.userParticipantRepo.GetByUserIDAndGroupID(ctx, tx, userID, group.ID)
		if err != nil {
			return domain.ErrForbidden
		}

		group.Name = name
		group.Description = description

		return s.groupRepo.Update(ctx, tx, group)
	})

	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *groupService) Delete(ctx context.Context, externalID uuid.UUID, userID int64) error {
	return database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		group, err := s.groupRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		_, err = s.userParticipantRepo.GetByUserIDAndGroupID(ctx, tx, userID, group.ID)
		if err != nil {
			return domain.ErrForbidden
		}

		return s.groupRepo.Revoke(ctx, tx, externalID)
	})
}
