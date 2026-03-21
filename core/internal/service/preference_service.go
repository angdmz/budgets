package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/repository"
)

type PreferenceService interface {
	Get(ctx context.Context, userID int64) (*domain.UserPreference, error)
	Update(ctx context.Context, userID int64, theme domain.Theme, language domain.Language, displayCurrency domain.Currency) (*domain.UserPreference, error)
}

type preferenceService struct {
	pool   *pgxpool.Pool
	prefRepo repository.UserPreferenceRepository
}

func NewPreferenceService(pool *pgxpool.Pool, prefRepo repository.UserPreferenceRepository) PreferenceService {
	return &preferenceService{
		pool:     pool,
		prefRepo: prefRepo,
	}
}

func (s *preferenceService) Get(ctx context.Context, userID int64) (*domain.UserPreference, error) {
	var pref *domain.UserPreference

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		pref, err = s.prefRepo.GetByUserID(ctx, tx, userID)
		if err != nil {
			// If not found, return defaults
			pref = &domain.UserPreference{
				UserID:          userID,
				Theme:           domain.ThemeLight,
				Language:        domain.LanguageEN,
				DisplayCurrency: domain.CurrencyUSD,
			}
			// Persist the defaults
			return s.prefRepo.Upsert(ctx, tx, pref)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return pref, nil
}

func (s *preferenceService) Update(ctx context.Context, userID int64, theme domain.Theme, language domain.Language, displayCurrency domain.Currency) (*domain.UserPreference, error) {
	if !theme.IsValid() {
		return nil, domain.ErrValidation
	}
	if !language.IsValid() {
		return nil, domain.ErrValidation
	}
	if !displayCurrency.IsValid() {
		return nil, domain.ErrValidation
	}

	var pref *domain.UserPreference

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		pref = &domain.UserPreference{
			UserID:          userID,
			Theme:           theme,
			Language:        language,
			DisplayCurrency: displayCurrency,
		}
		return s.prefRepo.Upsert(ctx, tx, pref)
	})

	if err != nil {
		return nil, err
	}

	return pref, nil
}
