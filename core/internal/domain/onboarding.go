package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OnboardingStatus represents the global onboarding state
type OnboardingStatus string

const (
	OnboardingInProgress OnboardingStatus = "in_progress"
	OnboardingCompleted  OnboardingStatus = "completed"
	OnboardingSkipped    OnboardingStatus = "skipped"
)

func (s OnboardingStatus) IsValid() bool {
	switch s {
	case OnboardingInProgress, OnboardingCompleted, OnboardingSkipped:
		return true
	}
	return false
}

// OnboardingStep represents a single step in the onboarding flow
type OnboardingStep string

const (
	StepWelcome               OnboardingStep = "welcome"
	StepChooseGroup           OnboardingStep = "choose_group"
	StepChooseCadence         OnboardingStep = "choose_cadence"
	StepAddExpectedExpenses   OnboardingStep = "add_expected_expenses"
	StepBudgetSummary         OnboardingStep = "budget_summary"
	StepRegisterActualExpense OnboardingStep = "register_actual_expense"
	StepCompareExpenses       OnboardingStep = "compare_expenses"
	StepDashboardTour         OnboardingStep = "dashboard_tour"
	StepComplete              OnboardingStep = "complete"
)

var onboardingStepOrder = []OnboardingStep{
	StepWelcome,
	StepChooseGroup,
	StepChooseCadence,
	StepAddExpectedExpenses,
	StepBudgetSummary,
	StepRegisterActualExpense,
	StepCompareExpenses,
	StepDashboardTour,
	StepComplete,
}

func (s OnboardingStep) IsValid() bool {
	for _, valid := range onboardingStepOrder {
		if s == valid {
			return true
		}
	}
	return false
}

func (s OnboardingStep) Next() (OnboardingStep, bool) {
	for i, step := range onboardingStepOrder {
		if step == s && i+1 < len(onboardingStepOrder) {
			return onboardingStepOrder[i+1], true
		}
	}
	return "", false
}

func (s OnboardingStep) Prev() (OnboardingStep, bool) {
	for i, step := range onboardingStepOrder {
		if step == s && i > 0 {
			return onboardingStepOrder[i-1], true
		}
	}
	return "", false
}

// OnboardingStepStatus represents the status of a single step
type OnboardingStepStatus string

const (
	StepPending   OnboardingStepStatus = "pending"
	StepCompleted OnboardingStepStatus = "completed"
	StepSkipped   OnboardingStepStatus = "skipped"
)

func (s OnboardingStepStatus) IsValid() bool {
	switch s {
	case StepPending, StepCompleted, StepSkipped:
		return true
	}
	return false
}

// AllOnboardingSteps returns the ordered list of onboarding steps
func AllOnboardingSteps() []OnboardingStep {
	return onboardingStepOrder
}

// --- Persistible (create) ---

type PersistibleUserOnboarding struct {
	userID int64
}

func NewPersistibleUserOnboarding(userID int64) (*PersistibleUserOnboarding, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: user_id must be positive", ErrValidation)
	}
	return &PersistibleUserOnboarding{userID: userID}, nil
}

func (o *PersistibleUserOnboarding) PersistTo(ctx context.Context, p Persister) (*PersistedUserOnboarding, error) {
	var ob PersistedUserOnboarding
	err := p.QueryRow(
		ctx,
		[]any{&ob.id, &ob.externalID, &ob.userID, &ob.status, &ob.currentStep, &ob.createdAt, &ob.updatedAt},
		`INSERT INTO user_onboardings (user_id, status, current_step)
		 VALUES ($1, 'in_progress', 'welcome')
		 RETURNING id, external_id, user_id, status, current_step, created_at, updated_at`,
		o.userID,
	)
	if err != nil {
		return nil, err
	}
	return &ob, nil
}

// --- Persisted (existing) ---

type PersistedUserOnboarding struct {
	id          int64
	externalID  uuid.UUID
	userID      int64
	status      OnboardingStatus
	currentStep OnboardingStep
	createdAt   time.Time
	updatedAt   time.Time
}

func PersistedUserOnboardingFromPersistence(ctx context.Context, userID int64, p Persister) (*PersistedUserOnboarding, error) {
	var ob PersistedUserOnboarding
	err := p.QueryRow(
		ctx,
		[]any{&ob.id, &ob.externalID, &ob.userID, &ob.status, &ob.currentStep, &ob.createdAt, &ob.updatedAt},
		`SELECT id, external_id, user_id, status, current_step, created_at, updated_at
		 FROM user_onboardings WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: onboarding not found", ErrNotFound)
	}
	return &ob, nil
}

func (o *PersistedUserOnboarding) ID() int64             { return o.id }
func (o *PersistedUserOnboarding) ExternalID() uuid.UUID { return o.externalID }
func (o *PersistedUserOnboarding) UserID() int64         { return o.userID }
func (o *PersistedUserOnboarding) Status() OnboardingStatus { return o.status }
func (o *PersistedUserOnboarding) CurrentStep() OnboardingStep { return o.currentStep }
func (o *PersistedUserOnboarding) CreatedAt() time.Time  { return o.createdAt }
func (o *PersistedUserOnboarding) UpdatedAt() time.Time  { return o.updatedAt }

func (o *PersistedUserOnboarding) Steps(ctx context.Context, p Persister) ([]PersistedUserOnboardingStep, error) {
	steps := make([]PersistedUserOnboardingStep, 0)
	err := p.QueryRows(
		ctx,
		func() []any {
			var s PersistedUserOnboardingStep
			steps = append(steps, s)
			idx := len(steps) - 1
			return []any{
				&steps[idx].id, &steps[idx].externalID, &steps[idx].onboardingID,
				&steps[idx].step, &steps[idx].status, &steps[idx].rawData,
				&steps[idx].createdAt, &steps[idx].updatedAt,
			}
		},
		`SELECT id, external_id, user_onboarding_id, step, status, data, created_at, updated_at
		 FROM user_onboarding_steps
		 WHERE user_onboarding_id = $1 AND revoked_at IS NULL
		 ORDER BY (CASE step
		   WHEN 'welcome' THEN 0
		   WHEN 'choose_group' THEN 1
		   WHEN 'choose_cadence' THEN 2
		   WHEN 'add_expected_expenses' THEN 3
		   WHEN 'budget_summary' THEN 4
		   WHEN 'register_actual_expense' THEN 5
		   WHEN 'compare_expenses' THEN 6
		   WHEN 'dashboard_tour' THEN 7
		   WHEN 'complete' THEN 8
		 END)`,
		o.id,
	)
	if err != nil {
		return nil, err
	}
	return steps, nil
}

func (o *PersistedUserOnboarding) StepByType(ctx context.Context, step OnboardingStep, p Persister) (*PersistedUserOnboardingStep, error) {
	var s PersistedUserOnboardingStep
	err := p.QueryRow(
		ctx,
		[]any{&s.id, &s.externalID, &s.onboardingID, &s.step, &s.status, &s.rawData, &s.createdAt, &s.updatedAt},
		`SELECT id, external_id, user_onboarding_id, step, status, data, created_at, updated_at
		 FROM user_onboarding_steps
		 WHERE user_onboarding_id = $1 AND step = $2 AND revoked_at IS NULL`,
		o.id, step,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: step not found", ErrNotFound)
	}
	return &s, nil
}

func (o *PersistedUserOnboarding) AdvanceTo(ctx context.Context, step OnboardingStep, p Persister) error {
	_, err := p.Exec(
		ctx,
		`UPDATE user_onboardings SET current_step = $2, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $1 AND revoked_at IS NULL`,
		o.id, step,
	)
	if err != nil {
		return err
	}
	o.currentStep = step
	return nil
}

func (o *PersistedUserOnboarding) MarkCompleted(ctx context.Context, p Persister) error {
	_, err := p.Exec(
		ctx,
		`UPDATE user_onboardings SET status = 'completed', current_step = 'complete', updated_at = CURRENT_TIMESTAMP
		 WHERE id = $1 AND revoked_at IS NULL`,
		o.id,
	)
	if err != nil {
		return err
	}
	o.status = OnboardingCompleted
	o.currentStep = StepComplete
	return nil
}

func (o *PersistedUserOnboarding) MarkSkipped(ctx context.Context, p Persister) error {
	_, err := p.Exec(
		ctx,
		`UPDATE user_onboardings SET status = 'skipped', updated_at = CURRENT_TIMESTAMP
		 WHERE id = $1 AND revoked_at IS NULL`,
		o.id,
	)
	if err != nil {
		return err
	}
	o.status = OnboardingSkipped
	return nil
}

func (o *PersistedUserOnboarding) Reset(ctx context.Context, p Persister) error {
	_, err := p.Exec(
		ctx,
		`UPDATE user_onboarding_steps SET revoked_at = CURRENT_TIMESTAMP
		 WHERE user_onboarding_id = $1 AND revoked_at IS NULL`,
		o.id,
	)
	if err != nil {
		return err
	}
	_, err = p.Exec(
		ctx,
		`UPDATE user_onboardings SET status = 'in_progress', current_step = 'welcome', updated_at = CURRENT_TIMESTAMP
		 WHERE id = $1 AND revoked_at IS NULL`,
		o.id,
	)
	if err != nil {
		return err
	}
	o.status = OnboardingInProgress
	o.currentStep = StepWelcome
	return nil
}

// --- Persistible step (create or upsert) ---

type PersistibleUserOnboardingStep struct {
	onboardingID int64
	step         OnboardingStep
	data         []byte
}

func NewPersistibleUserOnboardingStep(onboardingID int64, step OnboardingStep, data []byte) (*PersistibleUserOnboardingStep, error) {
	if onboardingID <= 0 {
		return nil, fmt.Errorf("%w: onboarding_id must be positive", ErrValidation)
	}
	if !step.IsValid() {
		return nil, fmt.Errorf("%w: invalid step", ErrValidation)
	}
	if err := validateStepData(step, data); err != nil {
		return nil, err
	}
	return &PersistibleUserOnboardingStep{
		onboardingID: onboardingID,
		step:         step,
		data:         data,
	}, nil
}

func (s *PersistibleUserOnboardingStep) PersistTo(ctx context.Context, p Persister) (*PersistedUserOnboardingStep, error) {
	var persisted PersistedUserOnboardingStep
	err := p.QueryRow(
		ctx,
		[]any{&persisted.id, &persisted.externalID, &persisted.onboardingID, &persisted.step, &persisted.status, &persisted.rawData, &persisted.createdAt, &persisted.updatedAt},
		`INSERT INTO user_onboarding_steps (user_onboarding_id, step, status, data)
		 VALUES ($1, $2, 'completed', $3)
		 ON CONFLICT (user_onboarding_id, step) WHERE revoked_at IS NULL
		 DO UPDATE SET status = 'completed', data = EXCLUDED.data, updated_at = CURRENT_TIMESTAMP
		 RETURNING id, external_id, user_onboarding_id, step, status, data, created_at, updated_at`,
		s.onboardingID, s.step, s.data,
	)
	if err != nil {
		return nil, err
	}
	return &persisted, nil
}

func (s *PersistibleUserOnboardingStep) PersistSkippedTo(ctx context.Context, p Persister) (*PersistedUserOnboardingStep, error) {
	var persisted PersistedUserOnboardingStep
	err := p.QueryRow(
		ctx,
		[]any{&persisted.id, &persisted.externalID, &persisted.onboardingID, &persisted.step, &persisted.status, &persisted.rawData, &persisted.createdAt, &persisted.updatedAt},
		`INSERT INTO user_onboarding_steps (user_onboarding_id, step, status, data)
		 VALUES ($1, $2, 'skipped', NULL)
		 ON CONFLICT (user_onboarding_id, step) WHERE revoked_at IS NULL
		 DO UPDATE SET status = 'skipped', data = NULL, updated_at = CURRENT_TIMESTAMP
		 RETURNING id, external_id, user_onboarding_id, step, status, data, created_at, updated_at`,
		s.onboardingID, s.step,
	)
	if err != nil {
		return nil, err
	}
	return &persisted, nil
}

// NewSkippedUserOnboardingStep creates a persistible step for skipping, without data validation.
func NewSkippedUserOnboardingStep(onboardingID int64, step OnboardingStep) (*PersistibleUserOnboardingStep, error) {
	if onboardingID <= 0 {
		return nil, fmt.Errorf("%w: onboarding_id must be positive", ErrValidation)
	}
	if !step.IsValid() {
		return nil, fmt.Errorf("%w: invalid step", ErrValidation)
	}
	return &PersistibleUserOnboardingStep{
		onboardingID: onboardingID,
		step:         step,
		data:         nil,
	}, nil
}

// --- Persisted step ---

type PersistedUserOnboardingStep struct {
	id           int64
	externalID   uuid.UUID
	onboardingID int64
	step         OnboardingStep
	status       OnboardingStepStatus
	rawData       []byte
	createdAt    time.Time
	updatedAt    time.Time
}

func (s *PersistedUserOnboardingStep) ID() int64              { return s.id }
func (s *PersistedUserOnboardingStep) ExternalID() uuid.UUID  { return s.externalID }
func (s *PersistedUserOnboardingStep) OnboardingID() int64    { return s.onboardingID }
func (s *PersistedUserOnboardingStep) Step() OnboardingStep   { return s.step }
func (s *PersistedUserOnboardingStep) Status() OnboardingStepStatus { return s.status }
func (s *PersistedUserOnboardingStep) Data() []byte           { return s.rawData }
func (s *PersistedUserOnboardingStep) CreatedAt() time.Time   { return s.createdAt }
func (s *PersistedUserOnboardingStep) UpdatedAt() time.Time   { return s.updatedAt }

// --- Step data validation ---

func validateStepData(step OnboardingStep, data []byte) error {
	if len(data) == 0 {
		data = []byte("{}")
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("%w: step data must be valid JSON", ErrValidation)
	}

	switch step {
	case StepWelcome:
		// No required fields
		return nil

	case StepChooseGroup:
		if _, ok := raw["group_id"]; !ok {
			return fmt.Errorf("%w: group_id is required", ErrValidation)
		}
		groupIDStr, ok := raw["group_id"].(string)
		if !ok {
			return fmt.Errorf("%w: group_id must be a string UUID", ErrValidation)
		}
		if _, err := uuid.Parse(groupIDStr); err != nil {
			return fmt.Errorf("%w: group_id must be a valid UUID", ErrValidation)
		}
		return nil

	case StepChooseCadence:
		if _, ok := raw["budget_id"]; !ok {
			return fmt.Errorf("%w: budget_id is required", ErrValidation)
		}
		budgetIDStr, ok := raw["budget_id"].(string)
		if !ok {
			return fmt.Errorf("%w: budget_id must be a string UUID", ErrValidation)
		}
		if _, err := uuid.Parse(budgetIDStr); err != nil {
			return fmt.Errorf("%w: budget_id must be a valid UUID", ErrValidation)
		}
		if cadence, ok := raw["cadence"]; ok {
			cadenceStr, ok := cadence.(string)
			if !ok {
				return fmt.Errorf("%w: cadence must be a string", ErrValidation)
			}
			validCadences := map[string]bool{
				"weekly": true, "biweekly": true, "monthly": true, "custom": true,
			}
			if !validCadences[cadenceStr] {
				return fmt.Errorf("%w: cadence must be one of weekly, biweekly, monthly, custom", ErrValidation)
			}
		}
		return nil

	case StepAddExpectedExpenses:
		expensesRaw, ok := raw["expected_expense_ids"]
		if !ok {
			return fmt.Errorf("%w: expected_expense_ids is required", ErrValidation)
		}
		expenses, ok := expensesRaw.([]interface{})
		if !ok {
			return fmt.Errorf("%w: expected_expense_ids must be an array", ErrValidation)
		}
		if len(expenses) == 0 {
			return fmt.Errorf("%w: expected_expense_ids must not be empty", ErrValidation)
		}
		for _, e := range expenses {
			eStr, ok := e.(string)
			if !ok {
				return fmt.Errorf("%w: each expected_expense_id must be a string UUID", ErrValidation)
			}
			if _, err := uuid.Parse(eStr); err != nil {
				return fmt.Errorf("%w: each expected_expense_id must be a valid UUID", ErrValidation)
			}
		}
		return nil

	case StepBudgetSummary:
		// Optional: total_expected and currency for display purposes
		return nil

	case StepRegisterActualExpense:
		if _, ok := raw["actual_expense_id"]; !ok {
			return fmt.Errorf("%w: actual_expense_id is required", ErrValidation)
		}
		expenseIDStr, ok := raw["actual_expense_id"].(string)
		if !ok {
			return fmt.Errorf("%w: actual_expense_id must be a string UUID", ErrValidation)
		}
		if _, err := uuid.Parse(expenseIDStr); err != nil {
			return fmt.Errorf("%w: actual_expense_id must be a valid UUID", ErrValidation)
		}
		return nil

	case StepCompareExpenses:
		// No required fields — comparison is derived
		return nil

	case StepDashboardTour:
		// No required fields
		return nil

	case StepComplete:
		// No required fields
		return nil
	}

	return fmt.Errorf("%w: unknown step", ErrValidation)
}
