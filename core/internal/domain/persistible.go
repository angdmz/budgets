package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/budgets/core/internal/representation"
)

type Persister interface {
	QueryRow(ctx context.Context, dest []any, query string, args ...any) error
	Exec(ctx context.Context, query string, args ...any) (int64, error)
	QueryRows(ctx context.Context, dest func() []any, query string, args ...any) error
}

type userParticipantData struct {
	user      *PersistedUser
	role      ParticipantRole
	isPrimary bool
}

type PersistibleGroup struct {
	name         string
	description  string
	participants []*PersistibleParticipant
}

func NewPersistibleGroup(name, description string) (*PersistibleGroup, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrValidation)
	}
	return &PersistibleGroup{
		name:         name,
		description:  description,
		participants: make([]*PersistibleParticipant, 0),
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

// AddParticipantForUser adds a participant named after the given user.
func (g *PersistibleGroup) AddParticipantForUser(user *PersistedUser) *PersistibleParticipant {
	return g.AddParticipant(user.displayName, "")
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
				userLink.user, participantID, string(userLink.role), isPrimaryInt,
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

func (p *PersistibleParticipant) AddPrimaryUser(user *PersistedUser, role ParticipantRole) {
	p.userLinks = append(p.userLinks, userParticipantData{
		user:      user,
		role:      role,
		isPrimary: true,
	})
}

func (p *PersistibleParticipant) AddMemberUser(user *PersistedUser, role ParticipantRole) {
	p.userLinks = append(p.userLinks, userParticipantData{
		user:      user,
		role:      role,
		isPrimary: false,
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
	var categoryID int64
	var categoryExternalID uuid.UUID
	var createdAt, updatedAt time.Time

	err := p.QueryRow(
		ctx,
		[]any{&categoryID, &categoryExternalID, &createdAt, &updatedAt},
		`INSERT INTO expense_categories (name, description, color, icon, budgeting_group_id)
		SELECT $1, $2, $3, $4, bg.id
		FROM budgeting_groups bg
		WHERE bg.external_id = $5 AND bg.revoked_at IS NULL
		RETURNING id, external_id, created_at, updated_at`,
		c.name, c.description, c.color, c.icon, c.groupExternalID,
	)
	if err != nil {
		return nil, wrapNotFound(err, "group not found")
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
	var budgetID int64
	var budgetExternalID uuid.UUID
	var createdAt, updatedAt time.Time

	err := p.QueryRow(
		ctx,
		[]any{&budgetID, &budgetExternalID, &createdAt, &updatedAt},
		`INSERT INTO budgets (name, description, start_date, end_date, budgeting_group_id)
		SELECT $1, $2, $3, $4, bg.id
		FROM budgeting_groups bg
		WHERE bg.external_id = $5 AND bg.revoked_at IS NULL
		RETURNING id, external_id, created_at, updated_at`,
		b.name, b.description, b.startDate, b.endDate, b.groupExternalID,
	)
	if err != nil {
		return nil, wrapNotFound(err, "group not found")
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
	categoryExternalID   uuid.UUID
}

func NewPersistibleExpectedExpense(name, description, encryptedAmount string, budgetExternalID uuid.UUID, categoryExternalID uuid.UUID) (*PersistibleExpectedExpense, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrValidation)
	}
	if encryptedAmount == "" {
		return nil, fmt.Errorf("%w: encrypted amount cannot be empty", ErrValidation)
	}
	if categoryExternalID == uuid.Nil {
		return nil, fmt.Errorf("%w: category cannot be empty", ErrValidation)
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
	var expenseID int64
	var expenseExternalID uuid.UUID
	var createdAt, updatedAt time.Time

	err := p.QueryRow(
		ctx,
		[]any{&expenseID, &expenseExternalID, &createdAt, &updatedAt},
		`INSERT INTO expected_expenses (name, description, encrypted_amount, budget_id, category_id)
		SELECT $1, $2, $3, b.id, ec.id
		FROM budgets b, expense_categories ec
		WHERE b.external_id = $4 AND b.revoked_at IS NULL
		AND ec.external_id = $5 AND ec.revoked_at IS NULL
		RETURNING id, external_id, created_at, updated_at`,
		e.name, e.description, e.encryptedAmount, e.budgetExternalID, e.categoryExternalID,
	)
	if err != nil {
		return nil, wrapNotFound(err, "budget or category not found")
	}

	return &PersistedExpectedExpense{
		id:                 expenseID,
		externalID:         expenseExternalID,
		name:               e.name,
		description:        e.description,
		encryptedAmount:    e.encryptedAmount,
		categoryExternalID: e.categoryExternalID,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
	}, nil
}

type PersistibleActualExpense struct {
	name                       string
	description                string
	expenseDate                time.Time
	encryptedAmount            string
	budgetExternalID           uuid.UUID
	categoryExternalID         uuid.UUID
	expectedExpenseExternalID  *uuid.UUID
}

func NewPersistibleActualExpense(name, description string, expenseDate time.Time, encryptedAmount string, budgetExternalID, categoryExternalID uuid.UUID, expectedExpenseExternalID *uuid.UUID) (*PersistibleActualExpense, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrValidation)
	}
	if encryptedAmount == "" {
		return nil, fmt.Errorf("%w: encrypted amount cannot be empty", ErrValidation)
	}
	if categoryExternalID == uuid.Nil {
		return nil, fmt.Errorf("%w: category cannot be empty", ErrValidation)
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
	var expenseID int64
	var expenseExternalID uuid.UUID
	var createdAt, updatedAt time.Time

	err := p.QueryRow(
		ctx,
		[]any{&expenseID, &expenseExternalID, &createdAt, &updatedAt},
		`INSERT INTO actual_expenses (name, description, expense_date, encrypted_amount, budget_id, category_id, expected_expense_id)
		SELECT $1, $2, $3, $4, b.id, ec.id, ee.id
		FROM budgets b
		CROSS JOIN expense_categories ec
		LEFT JOIN expected_expenses ee ON ee.external_id = $7 AND ee.revoked_at IS NULL
		WHERE b.external_id = $5 AND b.revoked_at IS NULL
		AND ec.external_id = $6 AND ec.revoked_at IS NULL
		AND ($7 IS NULL OR ee.id IS NOT NULL)
		RETURNING id, external_id, created_at, updated_at`,
		e.name, e.description, e.expenseDate, e.encryptedAmount, e.budgetExternalID, e.categoryExternalID, e.expectedExpenseExternalID,
	)
	if err != nil {
		return nil, wrapNotFound(err, "budget, category or expected expense not found")
	}

	return &PersistedActualExpense{
		id:                 expenseID,
		externalID:         expenseExternalID,
		name:               e.name,
		description:        e.description,
		expenseDate:        e.expenseDate,
		encryptedAmount:    e.encryptedAmount,
		categoryExternalID: e.categoryExternalID,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
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
		return nil, wrapNotFound(err, "group not found")
	}
	return &g, nil
}

// Render returns the final wire representation of this group.
func (g *PersistedGroup) Render() Rendered[representation.Group] {
	return render(representation.Group{
		ID:          g.externalID,
		Name:        g.name,
		Description: g.description,
		CreatedAt:   g.createdAt,
		UpdatedAt:   g.updatedAt,
	})
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
		return nil, wrapNotFound(err, "category not found")
	}
	return &c, nil
}

// Render returns the final wire representation of this category.
func (c *PersistedCategory) Render() Rendered[representation.Category] {
	return render(representation.Category{
		ID:          c.externalID,
		Name:        c.name,
		Description: c.description,
		Color:       c.color,
		Icon:        c.icon,
		CreatedAt:   c.createdAt,
		UpdatedAt:   c.updatedAt,
	})
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
		return nil, wrapNotFound(err, "budget not found")
	}
	return &b, nil
}

// conversionCutoffAt returns the as-of date for historical conversion of this
// budget's expected expenses: the earlier of the budget's end date and now.
func (b *PersistedBudget) conversionCutoffAt(now time.Time) time.Time {
	if b.endDate.After(now) {
		return now
	}
	return b.endDate
}

// Render returns the final wire representation of this budget.
func (b *PersistedBudget) Render() Rendered[representation.Budget] {
	return render(representation.Budget{
		ID:          b.externalID,
		Name:        b.name,
		Description: b.description,
		StartDate:   b.startDate.Format(apiDateFormat),
		EndDate:     b.endDate.Format(apiDateFormat),
		CreatedAt:   b.createdAt,
		UpdatedAt:   b.updatedAt,
	})
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
	id                 int64
	externalID         uuid.UUID
	name               string
	description        string
	encryptedAmount    string
	categoryExternalID uuid.UUID
	createdAt          time.Time
	updatedAt          time.Time
}

func PersistedExpectedExpenseFromPersistence(ctx context.Context, externalID uuid.UUID, p Persister) (*PersistedExpectedExpense, error) {
	var e PersistedExpectedExpense
	err := p.QueryRow(
		ctx,
		[]any{&e.id, &e.externalID, &e.name, &e.description, &e.encryptedAmount, &e.categoryExternalID, &e.createdAt, &e.updatedAt},
		`SELECT ee.id, ee.external_id, ee.name, ee.description, ee.encrypted_amount, ec.external_id, ee.created_at, ee.updated_at 
		FROM expected_expenses ee 
		JOIN expense_categories ec ON ee.category_id = ec.id AND ec.revoked_at IS NULL
		WHERE ee.external_id = $1 AND ee.revoked_at IS NULL`,
		externalID,
	)
	if err != nil {
		return nil, wrapNotFound(err, "expected expense not found")
	}
	return &e, nil
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

func (e *PersistedExpectedExpense) UpdateCategoryExternalID(categoryExternalID uuid.UUID) {
	e.categoryExternalID = categoryExternalID
}

func (e *PersistedExpectedExpense) UpdateIn(ctx context.Context, p Persister) error {
	var updatedAt time.Time
	err := p.QueryRow(
		ctx,
		[]any{&updatedAt},
		`UPDATE expected_expenses ee
		SET name = $1, description = $2, encrypted_amount = $3, category_id = ec.id, updated_at = CURRENT_TIMESTAMP
		FROM expense_categories ec
		WHERE ee.id = $4 AND ec.external_id = $5 AND ec.revoked_at IS NULL
		RETURNING ee.updated_at`,
		e.name, e.description, e.encryptedAmount, e.id, e.categoryExternalID,
	)
	if err != nil {
		return wrapNotFound(err, "category not found")
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
	id                 int64
	externalID         uuid.UUID
	name               string
	description        string
	expenseDate        time.Time
	encryptedAmount    string
	categoryExternalID uuid.UUID
	createdAt          time.Time
	updatedAt          time.Time
}

func PersistedActualExpenseFromPersistence(ctx context.Context, externalID uuid.UUID, p Persister) (*PersistedActualExpense, error) {
	var e PersistedActualExpense
	err := p.QueryRow(
		ctx,
		[]any{&e.id, &e.externalID, &e.name, &e.description, &e.expenseDate, &e.encryptedAmount, &e.categoryExternalID, &e.createdAt, &e.updatedAt},
		`SELECT ae.id, ae.external_id, ae.name, ae.description, ae.expense_date, ae.encrypted_amount, ec.external_id, ae.created_at, ae.updated_at 
		FROM actual_expenses ae 
		JOIN expense_categories ec ON ae.category_id = ec.id AND ec.revoked_at IS NULL
		WHERE ae.external_id = $1 AND ae.revoked_at IS NULL`,
		externalID,
	)
	if err != nil {
		return nil, wrapNotFound(err, "actual expense not found")
	}
	return &e, nil
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

func (e *PersistedActualExpense) UpdateCategoryExternalID(categoryExternalID uuid.UUID) {
	e.categoryExternalID = categoryExternalID
}

func (e *PersistedActualExpense) UpdateIn(ctx context.Context, p Persister) error {
	var updatedAt time.Time
	err := p.QueryRow(
		ctx,
		[]any{&updatedAt},
		`UPDATE actual_expenses ae
		SET name = $1, description = $2, expense_date = $3, encrypted_amount = $4, category_id = ec.id, updated_at = CURRENT_TIMESTAMP
		FROM expense_categories ec
		WHERE ae.id = $5 AND ec.external_id = $6 AND ec.revoked_at IS NULL
		RETURNING ae.updated_at`,
		e.name, e.description, e.expenseDate, e.encryptedAmount, e.id, e.categoryExternalID,
	)
	if err != nil {
		return wrapNotFound(err, "category not found")
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

func PersistedGroupsForUser(ctx context.Context, user *PersistedUser, p Persister) ([]PersistedGroup, error) {
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
		user,
	)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func PersistedCategoriesForGroup(ctx context.Context, groupExternalID uuid.UUID, p Persister) ([]PersistedCategory, error) {
	categories := make([]PersistedCategory, 0)
	err := p.QueryRows(
		ctx,
		func() []any {
			var c PersistedCategory
			categories = append(categories, c)
			idx := len(categories) - 1
			return []any{&categories[idx].id, &categories[idx].externalID, &categories[idx].name, &categories[idx].description, &categories[idx].color, &categories[idx].icon, &categories[idx].createdAt, &categories[idx].updatedAt}
		},
		`SELECT ec.id, ec.external_id, ec.name, ec.description, ec.color, ec.icon, ec.created_at, ec.updated_at
		FROM expense_categories ec
		JOIN budgeting_groups bg ON ec.budgeting_group_id = bg.id
		WHERE bg.external_id = $1 AND bg.revoked_at IS NULL AND ec.revoked_at IS NULL
		ORDER BY ec.created_at DESC`,
		groupExternalID,
	)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func PersistedBudgetsForGroup(ctx context.Context, groupExternalID uuid.UUID, p Persister) ([]PersistedBudget, error) {
	budgets := make([]PersistedBudget, 0)
	err := p.QueryRows(
		ctx,
		func() []any {
			var b PersistedBudget
			budgets = append(budgets, b)
			idx := len(budgets) - 1
			return []any{&budgets[idx].id, &budgets[idx].externalID, &budgets[idx].name, &budgets[idx].description, &budgets[idx].startDate, &budgets[idx].endDate, &budgets[idx].createdAt, &budgets[idx].updatedAt}
		},
		`SELECT b.id, b.external_id, b.name, b.description, b.start_date, b.end_date, b.created_at, b.updated_at
		FROM budgets b
		JOIN budgeting_groups bg ON b.budgeting_group_id = bg.id
		WHERE bg.external_id = $1 AND bg.revoked_at IS NULL AND b.revoked_at IS NULL
		ORDER BY b.created_at DESC`,
		groupExternalID,
	)
	if err != nil {
		return nil, err
	}
	return budgets, nil
}

func PersistedExpectedExpensesForBudget(ctx context.Context, budgetExternalID uuid.UUID, p Persister) ([]PersistedExpectedExpense, error) {
	expenses := make([]PersistedExpectedExpense, 0)
	err := p.QueryRows(
		ctx,
		func() []any {
			var e PersistedExpectedExpense
			expenses = append(expenses, e)
			idx := len(expenses) - 1
			return []any{&expenses[idx].id, &expenses[idx].externalID, &expenses[idx].name, &expenses[idx].description, &expenses[idx].encryptedAmount, &expenses[idx].categoryExternalID, &expenses[idx].createdAt, &expenses[idx].updatedAt}
		},
		`SELECT ee.id, ee.external_id, ee.name, ee.description, ee.encrypted_amount, ec.external_id, ee.created_at, ee.updated_at
		FROM expected_expenses ee
		JOIN expense_categories ec ON ee.category_id = ec.id AND ec.revoked_at IS NULL
		JOIN budgets b ON ee.budget_id = b.id AND b.revoked_at IS NULL
		WHERE b.external_id = $1 AND ee.revoked_at IS NULL
		ORDER BY ee.created_at DESC`,
		budgetExternalID,
	)
	if err != nil {
		return nil, err
	}
	return expenses, nil
}

type PersistibleUserPreference struct {
	user               *PersistedUser
	theme              Theme
	language           Language
	displayCurrency    Currency
	preferredQuoteType QuoteType
}

func NewPersistibleUserPreference(user *PersistedUser, theme Theme, language Language, displayCurrency Currency, preferredQuoteType QuoteType) (*PersistibleUserPreference, error) {
	if user == nil {
		return nil, fmt.Errorf("%w: user is required", ErrValidation)
	}
	if !theme.IsValid() {
		return nil, fmt.Errorf("%w: invalid theme", ErrValidation)
	}
	if !language.IsValid() {
		return nil, fmt.Errorf("%w: invalid language", ErrValidation)
	}
	if !displayCurrency.IsValid() {
		return nil, fmt.Errorf("%w: invalid currency", ErrValidation)
	}
	if !preferredQuoteType.IsValid() {
		return nil, fmt.Errorf("%w: invalid quote type", ErrValidation)
	}
	return &PersistibleUserPreference{
		user:               user,
		theme:              theme,
		language:           language,
		displayCurrency:    displayCurrency,
		preferredQuoteType: preferredQuoteType,
	}, nil
}

func (pref *PersistibleUserPreference) PersistTo(ctx context.Context, p Persister) (*PersistedUserPreference, error) {
	var id int64
	var externalID uuid.UUID
	var createdAt, updatedAt time.Time

	err := p.QueryRow(
		ctx,
		[]any{&id, &externalID, &createdAt, &updatedAt},
		`INSERT INTO user_preferences (user_id, theme, language, display_currency, preferred_quote_type)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id) WHERE revoked_at IS NULL
		 DO UPDATE SET theme = EXCLUDED.theme, language = EXCLUDED.language, display_currency = EXCLUDED.display_currency, preferred_quote_type = EXCLUDED.preferred_quote_type, updated_at = CURRENT_TIMESTAMP
		 RETURNING id, external_id, created_at, updated_at`,
		pref.user, pref.theme, pref.language, pref.displayCurrency, pref.preferredQuoteType,
	)
	if err != nil {
		return nil, err
	}

	return &PersistedUserPreference{
		id:                id,
		externalID:        externalID,
		userID:            pref.user.id,
		theme:             pref.theme,
		language:          pref.language,
		displayCurrency:   pref.displayCurrency,
		preferredQuoteType: pref.preferredQuoteType,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}, nil
}

type PersistedUserPreference struct {
	id                int64
	externalID        uuid.UUID
	userID            int64
	theme             Theme
	language          Language
	displayCurrency   Currency
	preferredQuoteType QuoteType
	createdAt         time.Time
	updatedAt         time.Time
}

func PersistedUserPreferenceFromPersistence(ctx context.Context, user *PersistedUser, p Persister) (*PersistedUserPreference, error) {
	var pref PersistedUserPreference
	err := p.QueryRow(
		ctx,
		[]any{&pref.id, &pref.externalID, &pref.theme, &pref.language, &pref.displayCurrency, &pref.preferredQuoteType, &pref.createdAt, &pref.updatedAt},
		`SELECT id, external_id, theme, language, display_currency, preferred_quote_type, created_at, updated_at
		 FROM user_preferences WHERE user_id = $1 AND revoked_at IS NULL`,
		user,
	)
	if err != nil {
		return nil, wrapNotFound(err, "user preferences not found")
	}
	pref.userID = user.id
	return &pref, nil
}

// PersistedUserPreferenceFor loads the user's preferences, creating and
// persisting a defaults row when none exists yet (find-or-create).
func PersistedUserPreferenceFor(ctx context.Context, user *PersistedUser, p Persister) (*PersistedUserPreference, error) {
	pref, err := PersistedUserPreferenceFromPersistence(ctx, user, p)
	if err == nil {
		return pref, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	persistible, err := NewPersistibleUserPreference(user, ThemeLight, LanguageEN, CurrencyUSD, QuoteOfficial)
	if err != nil {
		return nil, err
	}
	return persistible.PersistTo(ctx, p)
}

// PersistedUserPreferenceOrDefault loads the user's preferences without
// writing; when none are stored it returns an in-memory defaults object
// (NullObject) instead of an error.
func PersistedUserPreferenceOrDefault(ctx context.Context, user *PersistedUser, p Persister) (*PersistedUserPreference, error) {
	pref, err := PersistedUserPreferenceFromPersistence(ctx, user, p)
	if err == nil {
		return pref, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	return &PersistedUserPreference{
		userID:             user.id,
		theme:              ThemeLight,
		language:           LanguageEN,
		displayCurrency:    CurrencyUSD,
		preferredQuoteType: QuoteOfficial,
	}, nil
}

// Presentation returns the display settings used when converting and
// rendering amounts for this user.
func (pref *PersistedUserPreference) Presentation() PresentationPrefs {
	return PresentationPrefs{displayCurrency: pref.displayCurrency, preferredQuote: pref.preferredQuoteType}
}

// Render returns the final wire representation of these preferences.
func (pref *PersistedUserPreference) Render() Rendered[representation.Preference] {
	return render(representation.Preference{
		Theme:              string(pref.theme),
		Language:           string(pref.language),
		DisplayCurrency:    string(pref.displayCurrency),
		PreferredQuoteType: string(pref.preferredQuoteType),
	})
}

func (pref *PersistedUserPreference) UpdateTheme(theme Theme) error {
	if !theme.IsValid() {
		return fmt.Errorf("%w: invalid theme", ErrValidation)
	}
	pref.theme = theme
	return nil
}

func (pref *PersistedUserPreference) UpdateLanguage(language Language) error {
	if !language.IsValid() {
		return fmt.Errorf("%w: invalid language", ErrValidation)
	}
	pref.language = language
	return nil
}

func (pref *PersistedUserPreference) UpdateDisplayCurrency(currency Currency) error {
	if !currency.IsValid() {
		return fmt.Errorf("%w: invalid currency", ErrValidation)
	}
	pref.displayCurrency = currency
	return nil
}

func (pref *PersistedUserPreference) UpdatePreferredQuoteType(quoteType QuoteType) error {
	if !quoteType.IsValid() {
		return fmt.Errorf("%w: invalid quote type", ErrValidation)
	}
	pref.preferredQuoteType = quoteType
	return nil
}

func (pref *PersistedUserPreference) UpdateIn(ctx context.Context, p Persister) error {
	var updatedAt time.Time
	err := p.QueryRow(
		ctx,
		[]any{&updatedAt},
		`UPDATE user_preferences SET theme = $1, language = $2, display_currency = $3, preferred_quote_type = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5 RETURNING updated_at`,
		pref.theme, pref.language, pref.displayCurrency, pref.preferredQuoteType, pref.id,
	)
	if err != nil {
		return err
	}
	pref.updatedAt = updatedAt
	return nil
}

func PersistedActualExpensesForBudget(ctx context.Context, budgetExternalID uuid.UUID, p Persister) ([]PersistedActualExpense, error) {
	expenses := make([]PersistedActualExpense, 0)
	err := p.QueryRows(
		ctx,
		func() []any {
			var e PersistedActualExpense
			expenses = append(expenses, e)
			idx := len(expenses) - 1
			return []any{&expenses[idx].id, &expenses[idx].externalID, &expenses[idx].name, &expenses[idx].description, &expenses[idx].expenseDate, &expenses[idx].encryptedAmount, &expenses[idx].categoryExternalID, &expenses[idx].createdAt, &expenses[idx].updatedAt}
		},
		`SELECT ae.id, ae.external_id, ae.name, ae.description, ae.expense_date, ae.encrypted_amount, ec.external_id, ae.created_at, ae.updated_at
		FROM actual_expenses ae
		JOIN expense_categories ec ON ae.category_id = ec.id AND ec.revoked_at IS NULL
		JOIN budgets b ON ae.budget_id = b.id AND b.revoked_at IS NULL
		WHERE b.external_id = $1 AND ae.revoked_at IS NULL
		ORDER BY ae.created_at DESC`,
		budgetExternalID,
	)
	if err != nil {
		return nil, err
	}
	return expenses, nil
}
