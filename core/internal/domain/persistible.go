package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Persister interface {
	QueryRow(ctx context.Context, dest []any, query string, args ...any) error
	Exec(ctx context.Context, query string, args ...any) (int64, error)
	QueryRows(ctx context.Context, dest func() []any, query string, args ...any) error
}

type userParticipantData struct {
	userID    int64
	role      string
	isPrimary bool
}

type PersistibleGroup struct {
	name         string
	description  string
	participants []*PersistibleParticipant
	categories   []*PersistibleCategory
	budgets      []*PersistibleBudget
}

func NewPersistibleGroup(name, description string) (*PersistibleGroup, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrValidation)
	}
	return &PersistibleGroup{
		name:         name,
		description:  description,
		participants: make([]*PersistibleParticipant, 0),
		categories:   make([]*PersistibleCategory, 0),
		budgets:      make([]*PersistibleBudget, 0),
	}, nil
}

func (g *PersistibleGroup) AddParticipant(name, description string) *PersistibleParticipant {
	p := &PersistibleParticipant{
		name:        name,
		description: description,
		userLinks:   make([]userParticipantData, 0),
	}
	g.participants = append(g.participants, p)
	return p
}

func (g *PersistibleGroup) AddCategory(name, description, color, icon string) *PersistibleCategory {
	c := &PersistibleCategory{
		name:        name,
		description: description,
		color:       color,
		icon:        icon,
	}
	g.categories = append(g.categories, c)
	return c
}

func (g *PersistibleGroup) PersistTo(ctx context.Context, p Persister) (*PersistedGroup, error) {
	var groupID int64
	var groupExternalID uuid.UUID
	var createdAt, updatedAt time.Time

	err := p.QueryRow(
		ctx,
		[]any{&groupID, &groupExternalID, &createdAt, &updatedAt},
		`INSERT INTO budgeting_groups (name, description) VALUES ($1, $2) RETURNING id, external_id, created_at, updated_at`,
		g.name, g.description,
	)
	if err != nil {
		return nil, err
	}

	for _, participant := range g.participants {
		var participantID int64
		var participantExternalID uuid.UUID
		var pCreatedAt, pUpdatedAt time.Time

		err := p.QueryRow(
			ctx,
			[]any{&participantID, &participantExternalID, &pCreatedAt, &pUpdatedAt},
			`INSERT INTO participants (name, description, budgeting_group_id) VALUES ($1, $2, $3) RETURNING id, external_id, created_at, updated_at`,
			participant.name, participant.description, groupID,
		)
		if err != nil {
			return nil, err
		}

		for _, userLink := range participant.userLinks {
			isPrimaryInt := 0
			if userLink.isPrimary {
				isPrimaryInt = 1
			}
			_, err := p.Exec(
				ctx,
				`INSERT INTO user_participants (user_id, participant_id, role, is_primary) VALUES ($1, $2, $3, $4)`,
				userLink.userID, participantID, userLink.role, isPrimaryInt,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	return &PersistedGroup{
		id:          groupID,
		externalID:  groupExternalID,
		name:        g.name,
		description: g.description,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

type PersistibleParticipant struct {
	name        string
	description string
	userLinks   []userParticipantData
}

func (p *PersistibleParticipant) AddUser(userID int64, role string, isPrimary bool) {
	p.userLinks = append(p.userLinks, userParticipantData{
		userID:    userID,
		role:      role,
		isPrimary: isPrimary,
	})
}

type PersistibleCategory struct {
	name            string
	description     string
	color           string
	icon            string
	groupExternalID uuid.UUID
}

func NewPersistibleCategory(name, description, color, icon string, groupExternalID uuid.UUID) (*PersistibleCategory, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrValidation)
	}
	return &PersistibleCategory{
		name:            name,
		description:     description,
		color:           color,
		icon:            icon,
		groupExternalID: groupExternalID,
	}, nil
}

func (c *PersistibleCategory) PersistTo(ctx context.Context, p Persister) (*PersistedCategory, error) {
	var groupID int64
	err := p.QueryRow(
		ctx,
		[]any{&groupID},
		`SELECT id FROM budgeting_groups WHERE external_id = $1 AND revoked_at IS NULL`,
		c.groupExternalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: group not found", ErrNotFound)
	}

	var categoryID int64
	var categoryExternalID uuid.UUID
	var createdAt, updatedAt time.Time

	err = p.QueryRow(
		ctx,
		[]any{&categoryID, &categoryExternalID, &createdAt, &updatedAt},
		`INSERT INTO expense_categories (name, description, color, icon, budgeting_group_id) VALUES ($1, $2, $3, $4, $5) RETURNING id, external_id, created_at, updated_at`,
		c.name, c.description, c.color, c.icon, groupID,
	)
	if err != nil {
		return nil, err
	}

	return &PersistedCategory{
		id:          categoryID,
		externalID:  categoryExternalID,
		name:        c.name,
		description: c.description,
		color:       c.color,
		icon:        c.icon,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

type PersistibleBudget struct {
	name            string
	description     string
	startDate       time.Time
	endDate         time.Time
	groupExternalID uuid.UUID
}

func NewPersistibleBudget(name, description string, startDate, endDate time.Time, groupExternalID uuid.UUID) (*PersistibleBudget, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrValidation)
	}
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("%w: end date must be after start date", ErrValidation)
	}
	return &PersistibleBudget{
		name:            name,
		description:     description,
		startDate:       startDate,
		endDate:         endDate,
		groupExternalID: groupExternalID,
	}, nil
}

func (b *PersistibleBudget) PersistTo(ctx context.Context, p Persister) (*PersistedBudget, error) {
	var groupID int64
	err := p.QueryRow(
		ctx,
		[]any{&groupID},
		`SELECT id FROM budgeting_groups WHERE external_id = $1 AND revoked_at IS NULL`,
		b.groupExternalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: group not found", ErrNotFound)
	}

	var budgetID int64
	var budgetExternalID uuid.UUID
	var createdAt, updatedAt time.Time

	err = p.QueryRow(
		ctx,
		[]any{&budgetID, &budgetExternalID, &createdAt, &updatedAt},
		`INSERT INTO budgets (name, description, start_date, end_date, budgeting_group_id) VALUES ($1, $2, $3, $4, $5) RETURNING id, external_id, created_at, updated_at`,
		b.name, b.description, b.startDate, b.endDate, groupID,
	)
	if err != nil {
		return nil, err
	}

	return &PersistedBudget{
		id:          budgetID,
		externalID:  budgetExternalID,
		name:        b.name,
		description: b.description,
		startDate:   b.startDate,
		endDate:     b.endDate,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

type PersistibleExpectedExpense struct {
	name                 string
	description          string
	encryptedAmount      string
	budgetExternalID     uuid.UUID
	categoryExternalID   *uuid.UUID
}

func NewPersistibleExpectedExpense(name, description, encryptedAmount string, budgetExternalID uuid.UUID, categoryExternalID *uuid.UUID) (*PersistibleExpectedExpense, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrValidation)
	}
	if encryptedAmount == "" {
		return nil, fmt.Errorf("%w: encrypted amount cannot be empty", ErrValidation)
	}
	return &PersistibleExpectedExpense{
		name:               name,
		description:        description,
		encryptedAmount:    encryptedAmount,
		budgetExternalID:   budgetExternalID,
		categoryExternalID: categoryExternalID,
	}, nil
}

func (e *PersistibleExpectedExpense) PersistTo(ctx context.Context, p Persister) (*PersistedExpectedExpense, error) {
	var budgetID int64
	err := p.QueryRow(
		ctx,
		[]any{&budgetID},
		`SELECT id FROM budgets WHERE external_id = $1 AND revoked_at IS NULL`,
		e.budgetExternalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: budget not found", ErrNotFound)
	}

	var categoryID *int64
	if e.categoryExternalID != nil {
		var catID int64
		err := p.QueryRow(
			ctx,
			[]any{&catID},
			`SELECT id FROM expense_categories WHERE external_id = $1 AND revoked_at IS NULL`,
			*e.categoryExternalID,
		)
		if err != nil {
			return nil, fmt.Errorf("%w: category not found", ErrNotFound)
		}
		categoryID = &catID
	}

	var expenseID int64
	var expenseExternalID uuid.UUID
	var createdAt, updatedAt time.Time

	err = p.QueryRow(
		ctx,
		[]any{&expenseID, &expenseExternalID, &createdAt, &updatedAt},
		`INSERT INTO expected_expenses (name, description, encrypted_amount, budget_id, category_id) VALUES ($1, $2, $3, $4, $5) RETURNING id, external_id, created_at, updated_at`,
		e.name, e.description, e.encryptedAmount, budgetID, categoryID,
	)
	if err != nil {
		return nil, err
	}

	return &PersistedExpectedExpense{
		id:              expenseID,
		externalID:      expenseExternalID,
		name:            e.name,
		description:     e.description,
		encryptedAmount: e.encryptedAmount,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}, nil
}

type PersistibleActualExpense struct {
	name                       string
	description                string
	expenseDate                time.Time
	encryptedAmount            string
	budgetExternalID           uuid.UUID
	categoryExternalID         *uuid.UUID
	expectedExpenseExternalID  *uuid.UUID
}

func NewPersistibleActualExpense(name, description string, expenseDate time.Time, encryptedAmount string, budgetExternalID uuid.UUID, categoryExternalID, expectedExpenseExternalID *uuid.UUID) (*PersistibleActualExpense, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrValidation)
	}
	if encryptedAmount == "" {
		return nil, fmt.Errorf("%w: encrypted amount cannot be empty", ErrValidation)
	}
	return &PersistibleActualExpense{
		name:                      name,
		description:               description,
		expenseDate:               expenseDate,
		encryptedAmount:           encryptedAmount,
		budgetExternalID:          budgetExternalID,
		categoryExternalID:        categoryExternalID,
		expectedExpenseExternalID: expectedExpenseExternalID,
	}, nil
}

func (e *PersistibleActualExpense) PersistTo(ctx context.Context, p Persister) (*PersistedActualExpense, error) {
	var budgetID int64
	err := p.QueryRow(
		ctx,
		[]any{&budgetID},
		`SELECT id FROM budgets WHERE external_id = $1 AND revoked_at IS NULL`,
		e.budgetExternalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: budget not found", ErrNotFound)
	}

	var categoryID *int64
	if e.categoryExternalID != nil {
		var catID int64
		err := p.QueryRow(
			ctx,
			[]any{&catID},
			`SELECT id FROM expense_categories WHERE external_id = $1 AND revoked_at IS NULL`,
			*e.categoryExternalID,
		)
		if err != nil {
			return nil, fmt.Errorf("%w: category not found", ErrNotFound)
		}
		categoryID = &catID
	}

	var expectedExpenseID *int64
	if e.expectedExpenseExternalID != nil {
		var expID int64
		err := p.QueryRow(
			ctx,
			[]any{&expID},
			`SELECT id FROM expected_expenses WHERE external_id = $1 AND revoked_at IS NULL`,
			*e.expectedExpenseExternalID,
		)
		if err != nil {
			return nil, fmt.Errorf("%w: expected expense not found", ErrNotFound)
		}
		expectedExpenseID = &expID
	}

	var expenseID int64
	var expenseExternalID uuid.UUID
	var createdAt, updatedAt time.Time

	err = p.QueryRow(
		ctx,
		[]any{&expenseID, &expenseExternalID, &createdAt, &updatedAt},
		`INSERT INTO actual_expenses (name, description, expense_date, encrypted_amount, budget_id, category_id, expected_expense_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, external_id, created_at, updated_at`,
		e.name, e.description, e.expenseDate, e.encryptedAmount, budgetID, categoryID, expectedExpenseID,
	)
	if err != nil {
		return nil, err
	}

	return &PersistedActualExpense{
		id:              expenseID,
		externalID:      expenseExternalID,
		name:            e.name,
		description:     e.description,
		expenseDate:     e.expenseDate,
		encryptedAmount: e.encryptedAmount,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}, nil
}

type PersistedGroup struct {
	id          int64
	externalID  uuid.UUID
	name        string
	description string
	createdAt   time.Time
	updatedAt   time.Time
}

func PersistedGroupFromPersistence(ctx context.Context, externalID uuid.UUID, p Persister) (*PersistedGroup, error) {
	var g PersistedGroup
	err := p.QueryRow(
		ctx,
		[]any{&g.id, &g.externalID, &g.name, &g.description, &g.createdAt, &g.updatedAt},
		`SELECT id, external_id, name, description, created_at, updated_at FROM budgeting_groups WHERE external_id = $1 AND revoked_at IS NULL`,
		externalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: group not found", ErrNotFound)
	}
	return &g, nil
}

func (g *PersistedGroup) ExternalID() uuid.UUID {
	return g.externalID
}

func (g *PersistedGroup) Name() string {
	return g.name
}

func (g *PersistedGroup) Description() string {
	return g.description
}

func (g *PersistedGroup) CreatedAt() time.Time {
	return g.createdAt
}

func (g *PersistedGroup) UpdatedAt() time.Time {
	return g.updatedAt
}

func (g *PersistedGroup) UpdateName(name string) {
	g.name = name
}

func (g *PersistedGroup) UpdateDescription(description string) {
	g.description = description
}

func (g *PersistedGroup) UpdateIn(ctx context.Context, p Persister) error {
	var updatedAt time.Time
	err := p.QueryRow(
		ctx,
		[]any{&updatedAt},
		`UPDATE budgeting_groups SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 RETURNING updated_at`,
		g.name, g.description, g.id,
	)
	if err != nil {
		return err
	}
	g.updatedAt = updatedAt
	return nil
}

func (g *PersistedGroup) DeleteFrom(ctx context.Context, p Persister) error {
	_, err := p.Exec(
		ctx,
		`UPDATE budgeting_groups SET revoked_at = CURRENT_TIMESTAMP WHERE id = $1`,
		g.id,
	)
	return err
}

type PersistedCategory struct {
	id          int64
	externalID  uuid.UUID
	name        string
	description string
	color       string
	icon        string
	createdAt   time.Time
	updatedAt   time.Time
}

func PersistedCategoryFromPersistence(ctx context.Context, externalID uuid.UUID, p Persister) (*PersistedCategory, error) {
	var c PersistedCategory
	err := p.QueryRow(
		ctx,
		[]any{&c.id, &c.externalID, &c.name, &c.description, &c.color, &c.icon, &c.createdAt, &c.updatedAt},
		`SELECT id, external_id, name, description, color, icon, created_at, updated_at FROM expense_categories WHERE external_id = $1 AND revoked_at IS NULL`,
		externalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: category not found", ErrNotFound)
	}
	return &c, nil
}

func (c *PersistedCategory) ExternalID() uuid.UUID {
	return c.externalID
}

func (c *PersistedCategory) Name() string {
	return c.name
}

func (c *PersistedCategory) Description() string {
	return c.description
}

func (c *PersistedCategory) Color() string {
	return c.color
}

func (c *PersistedCategory) Icon() string {
	return c.icon
}

func (c *PersistedCategory) CreatedAt() time.Time {
	return c.createdAt
}

func (c *PersistedCategory) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *PersistedCategory) UpdateName(name string) {
	c.name = name
}

func (c *PersistedCategory) UpdateDescription(description string) {
	c.description = description
}

func (c *PersistedCategory) UpdateColor(color string) {
	c.color = color
}

func (c *PersistedCategory) UpdateIcon(icon string) {
	c.icon = icon
}

func (c *PersistedCategory) UpdateIn(ctx context.Context, p Persister) error {
	var updatedAt time.Time
	err := p.QueryRow(
		ctx,
		[]any{&updatedAt},
		`UPDATE expense_categories SET name = $1, description = $2, color = $3, icon = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5 RETURNING updated_at`,
		c.name, c.description, c.color, c.icon, c.id,
	)
	if err != nil {
		return err
	}
	c.updatedAt = updatedAt
	return nil
}

func (c *PersistedCategory) DeleteFrom(ctx context.Context, p Persister) error {
	_, err := p.Exec(
		ctx,
		`UPDATE expense_categories SET revoked_at = CURRENT_TIMESTAMP WHERE id = $1`,
		c.id,
	)
	return err
}

type PersistedBudget struct {
	id          int64
	externalID  uuid.UUID
	name        string
	description string
	startDate   time.Time
	endDate     time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

func PersistedBudgetFromPersistence(ctx context.Context, externalID uuid.UUID, p Persister) (*PersistedBudget, error) {
	var b PersistedBudget
	err := p.QueryRow(
		ctx,
		[]any{&b.id, &b.externalID, &b.name, &b.description, &b.startDate, &b.endDate, &b.createdAt, &b.updatedAt},
		`SELECT id, external_id, name, description, start_date, end_date, created_at, updated_at FROM budgets WHERE external_id = $1 AND revoked_at IS NULL`,
		externalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: budget not found", ErrNotFound)
	}
	return &b, nil
}

func (b *PersistedBudget) ExternalID() uuid.UUID {
	return b.externalID
}

func (b *PersistedBudget) Name() string {
	return b.name
}

func (b *PersistedBudget) Description() string {
	return b.description
}

func (b *PersistedBudget) StartDate() time.Time {
	return b.startDate
}

func (b *PersistedBudget) EndDate() time.Time {
	return b.endDate
}

func (b *PersistedBudget) CreatedAt() time.Time {
	return b.createdAt
}

func (b *PersistedBudget) UpdatedAt() time.Time {
	return b.updatedAt
}

func (b *PersistedBudget) UpdateName(name string) {
	b.name = name
}

func (b *PersistedBudget) UpdateDescription(description string) {
	b.description = description
}

func (b *PersistedBudget) UpdateDates(startDate, endDate time.Time) error {
	if endDate.Before(startDate) {
		return fmt.Errorf("%w: end date must be after start date", ErrValidation)
	}
	b.startDate = startDate
	b.endDate = endDate
	return nil
}

func (b *PersistedBudget) UpdateIn(ctx context.Context, p Persister) error {
	var updatedAt time.Time
	err := p.QueryRow(
		ctx,
		[]any{&updatedAt},
		`UPDATE budgets SET name = $1, description = $2, start_date = $3, end_date = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5 RETURNING updated_at`,
		b.name, b.description, b.startDate, b.endDate, b.id,
	)
	if err != nil {
		return err
	}
	b.updatedAt = updatedAt
	return nil
}

func (b *PersistedBudget) DeleteFrom(ctx context.Context, p Persister) error {
	_, err := p.Exec(
		ctx,
		`UPDATE budgets SET revoked_at = CURRENT_TIMESTAMP WHERE id = $1`,
		b.id,
	)
	return err
}

type PersistedExpectedExpense struct {
	id              int64
	externalID      uuid.UUID
	name            string
	description     string
	encryptedAmount string
	createdAt       time.Time
	updatedAt       time.Time
}

func PersistedExpectedExpenseFromPersistence(ctx context.Context, externalID uuid.UUID, p Persister) (*PersistedExpectedExpense, error) {
	var e PersistedExpectedExpense
	err := p.QueryRow(
		ctx,
		[]any{&e.id, &e.externalID, &e.name, &e.description, &e.encryptedAmount, &e.createdAt, &e.updatedAt},
		`SELECT id, external_id, name, description, encrypted_amount, created_at, updated_at FROM expected_expenses WHERE external_id = $1 AND revoked_at IS NULL`,
		externalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: expected expense not found", ErrNotFound)
	}
	return &e, nil
}

func (e *PersistedExpectedExpense) ExternalID() uuid.UUID {
	return e.externalID
}

func (e *PersistedExpectedExpense) Name() string {
	return e.name
}

func (e *PersistedExpectedExpense) Description() string {
	return e.description
}

func (e *PersistedExpectedExpense) EncryptedAmount() string {
	return e.encryptedAmount
}

func (e *PersistedExpectedExpense) CreatedAt() time.Time {
	return e.createdAt
}

func (e *PersistedExpectedExpense) UpdatedAt() time.Time {
	return e.updatedAt
}

func (e *PersistedExpectedExpense) UpdateName(name string) {
	e.name = name
}

func (e *PersistedExpectedExpense) UpdateDescription(description string) {
	e.description = description
}

func (e *PersistedExpectedExpense) UpdateEncryptedAmount(encryptedAmount string) {
	e.encryptedAmount = encryptedAmount
}

func (e *PersistedExpectedExpense) UpdateIn(ctx context.Context, p Persister) error {
	var updatedAt time.Time
	err := p.QueryRow(
		ctx,
		[]any{&updatedAt},
		`UPDATE expected_expenses SET name = $1, description = $2, encrypted_amount = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4 RETURNING updated_at`,
		e.name, e.description, e.encryptedAmount, e.id,
	)
	if err != nil {
		return err
	}
	e.updatedAt = updatedAt
	return nil
}

func (e *PersistedExpectedExpense) DeleteFrom(ctx context.Context, p Persister) error {
	_, err := p.Exec(
		ctx,
		`UPDATE expected_expenses SET revoked_at = CURRENT_TIMESTAMP WHERE id = $1`,
		e.id,
	)
	return err
}

type PersistedActualExpense struct {
	id              int64
	externalID      uuid.UUID
	name            string
	description     string
	expenseDate     time.Time
	encryptedAmount string
	createdAt       time.Time
	updatedAt       time.Time
}

func PersistedActualExpenseFromPersistence(ctx context.Context, externalID uuid.UUID, p Persister) (*PersistedActualExpense, error) {
	var e PersistedActualExpense
	err := p.QueryRow(
		ctx,
		[]any{&e.id, &e.externalID, &e.name, &e.description, &e.expenseDate, &e.encryptedAmount, &e.createdAt, &e.updatedAt},
		`SELECT id, external_id, name, description, expense_date, encrypted_amount, created_at, updated_at FROM actual_expenses WHERE external_id = $1 AND revoked_at IS NULL`,
		externalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: actual expense not found", ErrNotFound)
	}
	return &e, nil
}

func (e *PersistedActualExpense) ExternalID() uuid.UUID {
	return e.externalID
}

func (e *PersistedActualExpense) Name() string {
	return e.name
}

func (e *PersistedActualExpense) Description() string {
	return e.description
}

func (e *PersistedActualExpense) ExpenseDate() time.Time {
	return e.expenseDate
}

func (e *PersistedActualExpense) EncryptedAmount() string {
	return e.encryptedAmount
}

func (e *PersistedActualExpense) CreatedAt() time.Time {
	return e.createdAt
}

func (e *PersistedActualExpense) UpdatedAt() time.Time {
	return e.updatedAt
}

func (e *PersistedActualExpense) UpdateName(name string) {
	e.name = name
}

func (e *PersistedActualExpense) UpdateDescription(description string) {
	e.description = description
}

func (e *PersistedActualExpense) UpdateExpenseDate(expenseDate time.Time) {
	e.expenseDate = expenseDate
}

func (e *PersistedActualExpense) UpdateEncryptedAmount(encryptedAmount string) {
	e.encryptedAmount = encryptedAmount
}

func (e *PersistedActualExpense) UpdateIn(ctx context.Context, p Persister) error {
	var updatedAt time.Time
	err := p.QueryRow(
		ctx,
		[]any{&updatedAt},
		`UPDATE actual_expenses SET name = $1, description = $2, expense_date = $3, encrypted_amount = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5 RETURNING updated_at`,
		e.name, e.description, e.expenseDate, e.encryptedAmount, e.id,
	)
	if err != nil {
		return err
	}
	e.updatedAt = updatedAt
	return nil
}

func (e *PersistedActualExpense) DeleteFrom(ctx context.Context, p Persister) error {
	_, err := p.Exec(
		ctx,
		`UPDATE actual_expenses SET revoked_at = CURRENT_TIMESTAMP WHERE id = $1`,
		e.id,
	)
	return err
}

type SecurityGuard interface {
	AuthorizeGroupAccess(ctx context.Context, p Persister, groupExternalID uuid.UUID) error
	AuthorizeGroupOwnership(ctx context.Context, p Persister, groupExternalID uuid.UUID) error
	AuthorizeBudgetAccess(ctx context.Context, p Persister, budgetExternalID uuid.UUID) error
	AuthorizeCategoryAccess(ctx context.Context, p Persister, categoryExternalID uuid.UUID) error
	AuthorizeExpenseAccess(ctx context.Context, p Persister, expenseExternalID uuid.UUID) error
}

type securityGuard struct {
	userID int64
}

func NewSecurityGuard(userID int64) SecurityGuard {
	return &securityGuard{userID: userID}
}

func (s *securityGuard) AuthorizeGroupAccess(ctx context.Context, p Persister, groupExternalID uuid.UUID) error {
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
		s.userID, groupExternalID,
	)
	if err != nil {
		return err
	}
	if !exists {
		return ErrForbidden
	}
	return nil
}

func (s *securityGuard) AuthorizeGroupOwnership(ctx context.Context, p Persister, groupExternalID uuid.UUID) error {
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
		s.userID, groupExternalID,
	)
	if err != nil {
		return err
	}
	if !isOwner {
		return ErrForbidden
	}
	return nil
}

func (s *securityGuard) AuthorizeBudgetAccess(ctx context.Context, p Persister, budgetExternalID uuid.UUID) error {
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
		return ErrNotFound
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
		s.userID, budgetExternalID,
	)
	if err != nil {
		return err
	}
	if !hasAccess {
		return ErrForbidden
	}
	return nil
}

func (s *securityGuard) AuthorizeCategoryAccess(ctx context.Context, p Persister, categoryExternalID uuid.UUID) error {
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
		return ErrNotFound
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
		s.userID, categoryExternalID,
	)
	if err != nil {
		return err
	}
	if !hasAccess {
		return ErrForbidden
	}
	return nil
}

func (s *securityGuard) AuthorizeExpenseAccess(ctx context.Context, p Persister, expenseExternalID uuid.UUID) error {
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
		return ErrNotFound
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
		s.userID, expenseExternalID,
	)
	if err != nil {
		return err
	}
	if !hasAccess {
		return ErrForbidden
	}
	return nil
}

func PersistedGroupsForUser(ctx context.Context, userID int64, p Persister) ([]PersistedGroup, error) {
	groups := make([]PersistedGroup, 0)
	err := p.QueryRows(
		ctx,
		func() []any {
			var g PersistedGroup
			groups = append(groups, g)
			idx := len(groups) - 1
			return []any{&groups[idx].id, &groups[idx].externalID, &groups[idx].name, &groups[idx].description, &groups[idx].createdAt, &groups[idx].updatedAt}
		},
		`SELECT DISTINCT bg.id, bg.external_id, bg.name, bg.description, bg.created_at, bg.updated_at
		FROM budgeting_groups bg
		JOIN participants pt ON pt.budgeting_group_id = bg.id
		JOIN user_participants up ON up.participant_id = pt.id
		WHERE up.user_id = $1
		AND bg.revoked_at IS NULL AND pt.revoked_at IS NULL AND up.revoked_at IS NULL
		ORDER BY bg.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func PersistedCategoriesForGroup(ctx context.Context, groupExternalID uuid.UUID, p Persister) ([]PersistedCategory, error) {
	var groupID int64
	err := p.QueryRow(
		ctx,
		[]any{&groupID},
		`SELECT id FROM budgeting_groups WHERE external_id = $1 AND revoked_at IS NULL`,
		groupExternalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: group not found", ErrNotFound)
	}

	categories := make([]PersistedCategory, 0)
	err = p.QueryRows(
		ctx,
		func() []any {
			var c PersistedCategory
			categories = append(categories, c)
			idx := len(categories) - 1
			return []any{&categories[idx].id, &categories[idx].externalID, &categories[idx].name, &categories[idx].description, &categories[idx].color, &categories[idx].icon, &categories[idx].createdAt, &categories[idx].updatedAt}
		},
		`SELECT id, external_id, name, description, color, icon, created_at, updated_at
		FROM expense_categories
		WHERE budgeting_group_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func PersistedBudgetsForGroup(ctx context.Context, groupExternalID uuid.UUID, p Persister) ([]PersistedBudget, error) {
	var groupID int64
	err := p.QueryRow(
		ctx,
		[]any{&groupID},
		`SELECT id FROM budgeting_groups WHERE external_id = $1 AND revoked_at IS NULL`,
		groupExternalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: group not found", ErrNotFound)
	}

	budgets := make([]PersistedBudget, 0)
	err = p.QueryRows(
		ctx,
		func() []any {
			var b PersistedBudget
			budgets = append(budgets, b)
			idx := len(budgets) - 1
			return []any{&budgets[idx].id, &budgets[idx].externalID, &budgets[idx].name, &budgets[idx].description, &budgets[idx].startDate, &budgets[idx].endDate, &budgets[idx].createdAt, &budgets[idx].updatedAt}
		},
		`SELECT id, external_id, name, description, start_date, end_date, created_at, updated_at
		FROM budgets
		WHERE budgeting_group_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	return budgets, nil
}

func PersistedExpectedExpensesForBudget(ctx context.Context, budgetExternalID uuid.UUID, p Persister) ([]PersistedExpectedExpense, error) {
	var budgetID int64
	err := p.QueryRow(
		ctx,
		[]any{&budgetID},
		`SELECT id FROM budgets WHERE external_id = $1 AND revoked_at IS NULL`,
		budgetExternalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: budget not found", ErrNotFound)
	}

	expenses := make([]PersistedExpectedExpense, 0)
	err = p.QueryRows(
		ctx,
		func() []any {
			var e PersistedExpectedExpense
			expenses = append(expenses, e)
			idx := len(expenses) - 1
			return []any{&expenses[idx].id, &expenses[idx].externalID, &expenses[idx].name, &expenses[idx].description, &expenses[idx].encryptedAmount, &expenses[idx].createdAt, &expenses[idx].updatedAt}
		},
		`SELECT id, external_id, name, description, encrypted_amount, created_at, updated_at
		FROM expected_expenses
		WHERE budget_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC`,
		budgetID,
	)
	if err != nil {
		return nil, err
	}
	return expenses, nil
}

func PersistedActualExpensesForBudget(ctx context.Context, budgetExternalID uuid.UUID, p Persister) ([]PersistedActualExpense, error) {
	var budgetID int64
	err := p.QueryRow(
		ctx,
		[]any{&budgetID},
		`SELECT id FROM budgets WHERE external_id = $1 AND revoked_at IS NULL`,
		budgetExternalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: budget not found", ErrNotFound)
	}

	expenses := make([]PersistedActualExpense, 0)
	err = p.QueryRows(
		ctx,
		func() []any {
			var e PersistedActualExpense
			expenses = append(expenses, e)
			idx := len(expenses) - 1
			return []any{&expenses[idx].id, &expenses[idx].externalID, &expenses[idx].name, &expenses[idx].description, &expenses[idx].expenseDate, &expenses[idx].encryptedAmount, &expenses[idx].createdAt, &expenses[idx].updatedAt}
		},
		`SELECT id, external_id, name, description, expense_date, encrypted_amount, created_at, updated_at
		FROM actual_expenses
		WHERE budget_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC`,
		budgetID,
	)
	if err != nil {
		return nil, err
	}
	return expenses, nil
}
