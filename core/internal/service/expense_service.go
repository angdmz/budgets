package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/encryption"
	"github.com/budgets/core/internal/repository"
)

type ExpenseService interface {
	CreateExpected(ctx context.Context, budgetExternalID uuid.UUID, name, description string, amount decimal.Decimal, currency string, categoryExternalID *uuid.UUID, userID string) (*domain.ExpectedExpense, error)
	GetExpectedByID(ctx context.Context, externalID uuid.UUID, userID string) (*domain.ExpectedExpense, error)
	GetExpectedByBudgetID(ctx context.Context, budgetExternalID uuid.UUID, userID string) ([]domain.ExpectedExpense, error)
	UpdateExpected(ctx context.Context, externalID uuid.UUID, name, description string, amount decimal.Decimal, currency string, categoryExternalID *uuid.UUID, userID string) (*domain.ExpectedExpense, error)
	DeleteExpected(ctx context.Context, externalID uuid.UUID, userID string) error

	CreateActual(ctx context.Context, budgetExternalID uuid.UUID, name, description string, expenseDate time.Time, amount decimal.Decimal, currency string, categoryExternalID, expectedExpenseExternalID *uuid.UUID, userID string) (*domain.ActualExpense, error)
	GetActualByID(ctx context.Context, externalID uuid.UUID, userID string) (*domain.ActualExpense, error)
	GetActualByBudgetID(ctx context.Context, budgetExternalID uuid.UUID, userID string) ([]domain.ActualExpense, error)
	UpdateActual(ctx context.Context, externalID uuid.UUID, name, description string, expenseDate time.Time, amount decimal.Decimal, currency string, categoryExternalID, expectedExpenseExternalID *uuid.UUID, userID string) (*domain.ActualExpense, error)
	DeleteActual(ctx context.Context, externalID uuid.UUID, userID string) error
}

type expenseService struct {
	pool                *pgxpool.Pool
	encryptor           *encryption.Encryptor
	expectedExpenseRepo repository.ExpectedExpenseRepository
	actualExpenseRepo   repository.ActualExpenseRepository
	budgetRepo          repository.BudgetRepository
	categoryRepo        repository.ExpenseCategoryRepository
	participantRepo     repository.ParticipantRepository
}

func NewExpenseService(
	pool *pgxpool.Pool,
	encryptor *encryption.Encryptor,
	expectedExpenseRepo repository.ExpectedExpenseRepository,
	actualExpenseRepo repository.ActualExpenseRepository,
	budgetRepo repository.BudgetRepository,
	categoryRepo repository.ExpenseCategoryRepository,
	participantRepo repository.ParticipantRepository,
) ExpenseService {
	return &expenseService{
		pool:                pool,
		encryptor:           encryptor,
		expectedExpenseRepo: expectedExpenseRepo,
		actualExpenseRepo:   actualExpenseRepo,
		budgetRepo:          budgetRepo,
		categoryRepo:        categoryRepo,
		participantRepo:     participantRepo,
	}
}

func (s *expenseService) CreateExpected(ctx context.Context, budgetExternalID uuid.UUID, name, description string, amount decimal.Decimal, currency string, categoryExternalID *uuid.UUID, userID string) (*domain.ExpectedExpense, error) {
	var expense *domain.ExpectedExpense

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		budget, err := s.budgetRepo.GetByExternalID(ctx, tx, budgetExternalID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, budget.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		var categoryID *int64
		if categoryExternalID != nil {
			category, err := s.categoryRepo.GetByExternalID(ctx, tx, *categoryExternalID)
			if err != nil {
				return err
			}
			categoryID = &category.ID
		}

		money := encryption.Money{Amount: amount, Currency: currency}
		encryptedAmount, err := s.encryptor.EncryptMoney(money)
		if err != nil {
			return err
		}

		expense = &domain.ExpectedExpense{
			Name:        name,
			Description: description,
			Amount:      domain.Money{Amount: amount, Currency: domain.Currency(currency)},
			BudgetID:    budget.ID,
			CategoryID:  categoryID,
		}

		return s.expectedExpenseRepo.Create(ctx, tx, expense, encryptedAmount)
	})

	if err != nil {
		return nil, err
	}

	return expense, nil
}

func (s *expenseService) GetExpectedByID(ctx context.Context, externalID uuid.UUID, userID string) (*domain.ExpectedExpense, error) {
	var expense *domain.ExpectedExpense

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var encryptedAmount string
		var err error
		expense, encryptedAmount, err = s.expectedExpenseRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		budget, err := s.budgetRepo.GetByExternalID(ctx, tx, expense.BaseModel.ExternalID)
		if err != nil {
			budget2, _ := s.getBudgetByInternalID(ctx, tx, expense.BudgetID)
			if budget2 != nil {
				budget = budget2
			} else {
				return err
			}
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, budget.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		money, err := s.encryptor.DecryptMoney(encryptedAmount)
		if err != nil {
			return err
		}
		expense.Amount = domain.Money{Amount: money.Amount, Currency: domain.Currency(money.Currency)}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return expense, nil
}

func (s *expenseService) getBudgetByInternalID(ctx context.Context, tx pgx.Tx, budgetID int64) (*domain.Budget, error) {
	query := `SELECT id, external_id, name, description, start_date, end_date, budgeting_group_id, created_at, updated_at, revoked_at FROM budgets WHERE id = $1 AND revoked_at IS NULL`
	budget := &domain.Budget{}
	err := tx.QueryRow(ctx, query, budgetID).Scan(
		&budget.ID, &budget.ExternalID, &budget.Name, &budget.Description,
		&budget.StartDate, &budget.EndDate, &budget.BudgetingGroupID,
		&budget.CreatedAt, &budget.UpdatedAt, &budget.RevokedAt,
	)
	if err != nil {
		return nil, err
	}
	return budget, nil
}

func (s *expenseService) GetExpectedByBudgetID(ctx context.Context, budgetExternalID uuid.UUID, userID string) ([]domain.ExpectedExpense, error) {
	var expenses []domain.ExpectedExpense

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		budget, err := s.budgetRepo.GetByExternalID(ctx, tx, budgetExternalID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, budget.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		rawExpenses, encryptedAmounts, err := s.expectedExpenseRepo.GetByBudgetID(ctx, tx, budget.ID)
		if err != nil {
			return err
		}

		for i, exp := range rawExpenses {
			money, err := s.encryptor.DecryptMoney(encryptedAmounts[i])
			if err != nil {
				return err
			}
			exp.Amount = domain.Money{Amount: money.Amount, Currency: domain.Currency(money.Currency)}
			expenses = append(expenses, exp)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return expenses, nil
}

func (s *expenseService) UpdateExpected(ctx context.Context, externalID uuid.UUID, name, description string, amount decimal.Decimal, currency string, categoryExternalID *uuid.UUID, userID string) (*domain.ExpectedExpense, error) {
	var expense *domain.ExpectedExpense

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		expense, _, err = s.expectedExpenseRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		budget, err := s.getBudgetByInternalID(ctx, tx, expense.BudgetID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, budget.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		var categoryID *int64
		if categoryExternalID != nil {
			category, err := s.categoryRepo.GetByExternalID(ctx, tx, *categoryExternalID)
			if err != nil {
				return err
			}
			categoryID = &category.ID
		}

		money := encryption.Money{Amount: amount, Currency: currency}
		encryptedAmount, err := s.encryptor.EncryptMoney(money)
		if err != nil {
			return err
		}

		expense.Name = name
		expense.Description = description
		expense.Amount = domain.Money{Amount: amount, Currency: domain.Currency(currency)}
		expense.CategoryID = categoryID

		return s.expectedExpenseRepo.Update(ctx, tx, expense, encryptedAmount)
	})

	if err != nil {
		return nil, err
	}

	return expense, nil
}

func (s *expenseService) DeleteExpected(ctx context.Context, externalID uuid.UUID, userID string) error {
	return database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		expense, _, err := s.expectedExpenseRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		budget, err := s.getBudgetByInternalID(ctx, tx, expense.BudgetID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, budget.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		return s.expectedExpenseRepo.Revoke(ctx, tx, externalID)
	})
}

func (s *expenseService) CreateActual(ctx context.Context, budgetExternalID uuid.UUID, name, description string, expenseDate time.Time, amount decimal.Decimal, currency string, categoryExternalID, expectedExpenseExternalID *uuid.UUID, userID string) (*domain.ActualExpense, error) {
	var expense *domain.ActualExpense

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		budget, err := s.budgetRepo.GetByExternalID(ctx, tx, budgetExternalID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, budget.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		var categoryID *int64
		if categoryExternalID != nil {
			category, err := s.categoryRepo.GetByExternalID(ctx, tx, *categoryExternalID)
			if err != nil {
				return err
			}
			categoryID = &category.ID
		}

		var expectedExpenseID *int64
		if expectedExpenseExternalID != nil {
			expectedExpense, _, err := s.expectedExpenseRepo.GetByExternalID(ctx, tx, *expectedExpenseExternalID)
			if err != nil {
				return err
			}
			expectedExpenseID = &expectedExpense.ID
		}

		money := encryption.Money{Amount: amount, Currency: currency}
		encryptedAmount, err := s.encryptor.EncryptMoney(money)
		if err != nil {
			return err
		}

		expense = &domain.ActualExpense{
			Name:              name,
			Description:       description,
			ExpenseDate:       expenseDate,
			Amount:            domain.Money{Amount: amount, Currency: domain.Currency(currency)},
			BudgetID:          budget.ID,
			CategoryID:        categoryID,
			ExpectedExpenseID: expectedExpenseID,
		}

		return s.actualExpenseRepo.Create(ctx, tx, expense, encryptedAmount)
	})

	if err != nil {
		return nil, err
	}

	return expense, nil
}

func (s *expenseService) GetActualByID(ctx context.Context, externalID uuid.UUID, userID string) (*domain.ActualExpense, error) {
	var expense *domain.ActualExpense

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var encryptedAmount string
		var err error
		expense, encryptedAmount, err = s.actualExpenseRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		budget, err := s.getBudgetByInternalID(ctx, tx, expense.BudgetID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, budget.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		money, err := s.encryptor.DecryptMoney(encryptedAmount)
		if err != nil {
			return err
		}
		expense.Amount = domain.Money{Amount: money.Amount, Currency: domain.Currency(money.Currency)}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return expense, nil
}

func (s *expenseService) GetActualByBudgetID(ctx context.Context, budgetExternalID uuid.UUID, userID string) ([]domain.ActualExpense, error) {
	var expenses []domain.ActualExpense

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		budget, err := s.budgetRepo.GetByExternalID(ctx, tx, budgetExternalID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, budget.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		rawExpenses, encryptedAmounts, err := s.actualExpenseRepo.GetByBudgetID(ctx, tx, budget.ID)
		if err != nil {
			return err
		}

		for i, exp := range rawExpenses {
			money, err := s.encryptor.DecryptMoney(encryptedAmounts[i])
			if err != nil {
				return err
			}
			exp.Amount = domain.Money{Amount: money.Amount, Currency: domain.Currency(money.Currency)}
			expenses = append(expenses, exp)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return expenses, nil
}

func (s *expenseService) UpdateActual(ctx context.Context, externalID uuid.UUID, name, description string, expenseDate time.Time, amount decimal.Decimal, currency string, categoryExternalID, expectedExpenseExternalID *uuid.UUID, userID string) (*domain.ActualExpense, error) {
	var expense *domain.ActualExpense

	err := database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		expense, _, err = s.actualExpenseRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		budget, err := s.getBudgetByInternalID(ctx, tx, expense.BudgetID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, budget.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		var categoryID *int64
		if categoryExternalID != nil {
			category, err := s.categoryRepo.GetByExternalID(ctx, tx, *categoryExternalID)
			if err != nil {
				return err
			}
			categoryID = &category.ID
		}

		var expectedExpenseID *int64
		if expectedExpenseExternalID != nil {
			expectedExpense, _, err := s.expectedExpenseRepo.GetByExternalID(ctx, tx, *expectedExpenseExternalID)
			if err != nil {
				return err
			}
			expectedExpenseID = &expectedExpense.ID
		}

		money := encryption.Money{Amount: amount, Currency: currency}
		encryptedAmount, err := s.encryptor.EncryptMoney(money)
		if err != nil {
			return err
		}

		expense.Name = name
		expense.Description = description
		expense.ExpenseDate = expenseDate
		expense.Amount = domain.Money{Amount: amount, Currency: domain.Currency(currency)}
		expense.CategoryID = categoryID
		expense.ExpectedExpenseID = expectedExpenseID

		return s.actualExpenseRepo.Update(ctx, tx, expense, encryptedAmount)
	})

	if err != nil {
		return nil, err
	}

	return expense, nil
}

func (s *expenseService) DeleteActual(ctx context.Context, externalID uuid.UUID, userID string) error {
	return database.WithTransaction(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		expense, _, err := s.actualExpenseRepo.GetByExternalID(ctx, tx, externalID)
		if err != nil {
			return err
		}

		budget, err := s.getBudgetByInternalID(ctx, tx, expense.BudgetID)
		if err != nil {
			return err
		}

		_, err = s.participantRepo.GetByUserIDAndGroupID(ctx, tx, userID, budget.BudgetingGroupID)
		if err != nil {
			return domain.ErrForbidden
		}

		return s.actualExpenseRepo.Revoke(ctx, tx, externalID)
	})
}
