package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/budgets/core/internal/representation"
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
	user *PersistedUser
}

func NewPersistibleUserOnboarding(user *PersistedUser) (*PersistibleUserOnboarding, error) {
	if user == nil {
		return nil, fmt.Errorf("%w: user is required", ErrValidation)
	}
	return &PersistibleUserOnboarding{user: user}, nil
}

func (o *PersistibleUserOnboarding) PersistTo(ctx context.Context, p Persister) (*PersistedUserOnboarding, error) {
	var ob PersistedUserOnboarding
	err := p.QueryRow(
		ctx,
		[]any{&ob.id, &ob.externalID, &ob.userID, &ob.status, &ob.currentStep, &ob.createdAt, &ob.updatedAt},
		`INSERT INTO user_onboardings (user_id, status, current_step)
		 VALUES ($1, 'in_progress', 'welcome')
		 RETURNING id, external_id, user_id, status, current_step, created_at, updated_at`,
		o.user,
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

func PersistedUserOnboardingFromPersistence(ctx context.Context, user *PersistedUser, p Persister) (*PersistedUserOnboarding, error) {
	var ob PersistedUserOnboarding
	err := p.QueryRow(
		ctx,
		[]any{&ob.id, &ob.externalID, &ob.userID, &ob.status, &ob.currentStep, &ob.createdAt, &ob.updatedAt},
		`SELECT id, external_id, user_id, status, current_step, created_at, updated_at
		 FROM user_onboardings WHERE user_id = $1 AND revoked_at IS NULL`,
		user,
	)
	if err != nil {
		return nil, wrapNotFound(err, "onboarding not found")
	}
	return &ob, nil
}

// PersistedUserOnboardingFor loads the user's onboarding, creating and
// persisting a fresh one when none exists yet (find-or-create).
func PersistedUserOnboardingFor(ctx context.Context, user *PersistedUser, p Persister) (*PersistedUserOnboarding, error) {
	ob, err := PersistedUserOnboardingFromPersistence(ctx, user, p)
	if err == nil {
		return ob, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	persistible, err := NewPersistibleUserOnboarding(user)
	if err != nil {
		return nil, err
	}
	return persistible.PersistTo(ctx, p)
}

// Detail loads this onboarding's persisted steps and returns a renderable
// composite. The query happens here, at construction, so Render stays pure.
func (o *PersistedUserOnboarding) Detail(ctx context.Context, p Persister) (*OnboardingDetail, error) {
	steps, err := o.Steps(ctx, p)
	if err != nil {
		return nil, err
	}
	return &OnboardingDetail{onboarding: o, steps: steps}, nil
}

// NewCompletedStep builds a persistible completed step wired to this
// onboarding's internal id — the FK never leaves the domain package.
func (o *PersistedUserOnboarding) NewCompletedStep(step OnboardingStep, data []byte) (*PersistibleCompletedStep, error) {
	if !step.IsValid() {
		return nil, fmt.Errorf("%w: invalid step", ErrValidation)
	}
	if err := validateStepData(step, data); err != nil {
		return nil, err
	}
	return &PersistibleCompletedStep{
		onboardingID: o.id,
		step:         step,
		data:         data,
	}, nil
}

// NewSkippedStep builds a persistible skipped step wired to this onboarding's
// internal id.
func (o *PersistedUserOnboarding) NewSkippedStep(step OnboardingStep) (*PersistibleSkippedStep, error) {
	if !step.IsValid() {
		return nil, fmt.Errorf("%w: invalid step", ErrValidation)
	}
	return &PersistibleSkippedStep{
		onboardingID: o.id,
		step:         step,
	}, nil
}

// CompleteStep records `step` as completed, advances the flow (or marks the
// onboarding completed when `step` is the last one), and returns the refreshed
// renderable composite.
func (o *PersistedUserOnboarding) CompleteStep(ctx context.Context, step OnboardingStep, data []byte, p Persister) (*OnboardingDetail, error) {
	persistible, err := o.NewCompletedStep(step, data)
	if err != nil {
		return nil, err
	}
	if _, err := persistible.PersistTo(ctx, p); err != nil {
		return nil, err
	}
	if err := o.advanceOrComplete(ctx, step, p); err != nil {
		return nil, err
	}
	return o.Detail(ctx, p)
}

// SkipStep records `step` as skipped, advances the flow (or marks the
// onboarding completed when `step` is the last one), and returns the refreshed
// renderable composite.
func (o *PersistedUserOnboarding) SkipStep(ctx context.Context, step OnboardingStep, p Persister) (*OnboardingDetail, error) {
	persistible, err := o.NewSkippedStep(step)
	if err != nil {
		return nil, err
	}
	if _, err := persistible.PersistTo(ctx, p); err != nil {
		return nil, err
	}
	if err := o.advanceOrComplete(ctx, step, p); err != nil {
		return nil, err
	}
	return o.Detail(ctx, p)
}

// GoBackFrom moves the current step marker back to the step preceding `step`
// and returns the refreshed renderable composite.
func (o *PersistedUserOnboarding) GoBackFrom(ctx context.Context, step OnboardingStep, p Persister) (*OnboardingDetail, error) {
	if !step.IsValid() {
		return nil, fmt.Errorf("%w: invalid step", ErrValidation)
	}
	prev, ok := step.Prev()
	if !ok {
		return nil, fmt.Errorf("%w: cannot go back from the first step", ErrValidation)
	}
	if err := o.AdvanceTo(ctx, prev, p); err != nil {
		return nil, err
	}
	return o.Detail(ctx, p)
}

func (o *PersistedUserOnboarding) advanceOrComplete(ctx context.Context, step OnboardingStep, p Persister) error {
	if next, ok := step.Next(); ok {
		return o.AdvanceTo(ctx, next, p)
	}
	return o.MarkCompleted(ctx, p)
}

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
		return nil, wrapNotFound(err, "step not found")
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

// --- Persistible steps (create or upsert) ---

// PersistibleCompletedStep is a step-completion event to be persisted.
type PersistibleCompletedStep struct {
	onboardingID int64
	step         OnboardingStep
	data         []byte
}

func (s *PersistibleCompletedStep) PersistTo(ctx context.Context, p Persister) (*PersistedUserOnboardingStep, error) {
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

// PersistibleSkippedStep is a step-skip event to be persisted.
type PersistibleSkippedStep struct {
	onboardingID int64
	step         OnboardingStep
}

func (s *PersistibleSkippedStep) PersistTo(ctx context.Context, p Persister) (*PersistedUserOnboardingStep, error) {
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

// render produces this step's wire shape; unmarshalling the stored JSON data
// is pure computation, so it is safe to do at render time.
func (s *PersistedUserOnboardingStep) render() representation.OnboardingStep {
	var data interface{}
	if len(s.rawData) > 0 {
		_ = json.Unmarshal(s.rawData, &data)
	}
	return representation.OnboardingStep{
		Step:      string(s.step),
		Status:    string(s.status),
		Data:      data,
		CreatedAt: s.createdAt,
		UpdatedAt: s.updatedAt,
	}
}

// OnboardingDetail is a persisted onboarding composed with its steps; it
// renders the full onboarding wire shape.
type OnboardingDetail struct {
	onboarding *PersistedUserOnboarding
	steps      []PersistedUserOnboardingStep
}

// Render returns the final wire representation of the onboarding and its steps.
func (d *OnboardingDetail) Render() Rendered[representation.Onboarding] {
	steps := make([]representation.OnboardingStep, len(d.steps))
	for i := range d.steps {
		steps[i] = d.steps[i].render()
	}
	return render(representation.Onboarding{
		ID:          d.onboarding.externalID,
		Status:      string(d.onboarding.status),
		CurrentStep: string(d.onboarding.currentStep),
		Steps:       steps,
		CreatedAt:   d.onboarding.createdAt,
		UpdatedAt:   d.onboarding.updatedAt,
	})
}

// --- Step data validation ---

// stepDataValidator validates a step's decoded JSON payload. Polymorphism by
// table lookup instead of a switch: registering a new step means adding one
// entry here.
type stepDataValidator func(raw map[string]interface{}) error

func alwaysValid(map[string]interface{}) error { return nil }

func requireUUIDField(field string) stepDataValidator {
	return func(raw map[string]interface{}) error {
		value, ok := raw[field]
		if !ok {
			return fmt.Errorf("%w: %s is required", ErrValidation, field)
		}
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("%w: %s must be a string UUID", ErrValidation, field)
		}
		if _, err := uuid.Parse(str); err != nil {
			return fmt.Errorf("%w: %s must be a valid UUID", ErrValidation, field)
		}
		return nil
	}
}

func validateCadenceStep(raw map[string]interface{}) error {
	if err := requireUUIDField("budget_id")(raw); err != nil {
		return err
	}
	cadence, ok := raw["cadence"]
	if !ok {
		return nil
	}
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
	return nil
}

func validateExpectedExpensesStep(raw map[string]interface{}) error {
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
}

var stepValidators = map[OnboardingStep]stepDataValidator{
	StepWelcome:               alwaysValid,
	StepChooseGroup:           requireUUIDField("group_id"),
	StepChooseCadence:         validateCadenceStep,
	StepAddExpectedExpenses:   validateExpectedExpensesStep,
	StepBudgetSummary:         alwaysValid, // optional display fields
	StepRegisterActualExpense: requireUUIDField("actual_expense_id"),
	StepCompareExpenses:       alwaysValid, // comparison is derived
	StepDashboardTour:         alwaysValid,
	StepComplete:              alwaysValid,
}

func validateStepData(step OnboardingStep, data []byte) error {
	if len(data) == 0 {
		data = []byte("{}")
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("%w: step data must be valid JSON", ErrValidation)
	}

	validate, ok := stepValidators[step]
	if !ok {
		return fmt.Errorf("%w: unknown step", ErrValidation)
	}
	return validate(raw)
}
