package domain

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Repository is the interface that all repositories must implement
type Repository interface {
	// Transaction returns the current transaction or nil
	Transaction() pgx.Tx
}

// Persistible is the interface for domain entities that can be persisted
type Persistible interface {
	// PersistTo saves the entity using the provided repository
	PersistTo(ctx context.Context, repo Repository) error
}

// PersistibleGroup is a rich domain model for BudgetingGroup
type PersistibleGroup struct {
	BudgetingGroup
	participants []*PersistibleParticipant
	categories   []*PersistibleCategory
	budgets      []*PersistibleBudget
	isNew        bool
	isDirty      bool
}

func NewPersistibleGroup(name, description string) *PersistibleGroup {
	return &PersistibleGroup{
		BudgetingGroup: BudgetingGroup{
			Name:        name,
			Description: description,
		},
		participants: make([]*PersistibleParticipant, 0),
		categories:   make([]*PersistibleCategory, 0),
		budgets:      make([]*PersistibleBudget, 0),
		isNew:        true,
		isDirty:      true,
	}
}

func (g *PersistibleGroup) AddParticipant(name, description string) *PersistibleParticipant {
	p := &PersistibleParticipant{
		Participant: Participant{
			Name:        name,
			Description: description,
		},
		isNew:   true,
		isDirty: true,
	}
	g.participants = append(g.participants, p)
	return p
}

func (g *PersistibleGroup) AddCategory(name, description, color, icon string) *PersistibleCategory {
	c := &PersistibleCategory{
		ExpenseCategory: ExpenseCategory{
			Name:        name,
			Description: description,
			Color:       color,
			Icon:        icon,
		},
		isNew:   true,
		isDirty: true,
	}
	g.categories = append(g.categories, c)
	return c
}

func (g *PersistibleGroup) Participants() []*PersistibleParticipant {
	result := make([]*PersistibleParticipant, len(g.participants))
	copy(result, g.participants)
	return result
}

func (g *PersistibleGroup) Categories() []*PersistibleCategory {
	result := make([]*PersistibleCategory, len(g.categories))
	copy(result, g.categories)
	return result
}

func (g *PersistibleGroup) IsNew() bool {
	return g.isNew
}

func (g *PersistibleGroup) MarkPersisted() {
	g.isNew = false
	g.isDirty = false
}

// PersistibleParticipant is a rich domain model for Participant
type PersistibleParticipant struct {
	Participant
	userParticipants []*UserParticipant
	isNew            bool
	isDirty          bool
}

func (p *PersistibleParticipant) AddUser(userID int64, role string, isPrimary bool) {
	up := &UserParticipant{
		UserID:    userID,
		Role:      role,
		IsPrimary: isPrimary,
	}
	p.userParticipants = append(p.userParticipants, up)
	p.isDirty = true
}

func (p *PersistibleParticipant) UserParticipants() []*UserParticipant {
	result := make([]*UserParticipant, len(p.userParticipants))
	copy(result, p.userParticipants)
	return result
}

func (p *PersistibleParticipant) IsNew() bool {
	return p.isNew
}

func (p *PersistibleParticipant) MarkPersisted() {
	p.isNew = false
	p.isDirty = false
}

// PersistibleCategory is a rich domain model for ExpenseCategory
type PersistibleCategory struct {
	ExpenseCategory
	isNew   bool
	isDirty bool
}

func (c *PersistibleCategory) IsNew() bool {
	return c.isNew
}

func (c *PersistibleCategory) MarkPersisted() {
	c.isNew = false
	c.isDirty = false
}

func (c *PersistibleCategory) UpdateName(name string) {
	c.Name = name
	c.isDirty = true
}

func (c *PersistibleCategory) UpdateDescription(description string) {
	c.Description = description
	c.isDirty = true
}

// PersistibleBudget is a rich domain model for Budget
type PersistibleBudget struct {
	Budget
	expectedExpenses []*PersistibleExpectedExpense
	actualExpenses   []*PersistibleActualExpense
	isNew            bool
	isDirty          bool
}

func (b *PersistibleBudget) IsNew() bool {
	return b.isNew
}

func (b *PersistibleBudget) MarkPersisted() {
	b.isNew = false
	b.isDirty = false
}

// PersistibleExpectedExpense is a rich domain model for ExpectedExpense
type PersistibleExpectedExpense struct {
	ExpectedExpense
	isNew   bool
	isDirty bool
}

func (e *PersistibleExpectedExpense) IsNew() bool {
	return e.isNew
}

func (e *PersistibleExpectedExpense) MarkPersisted() {
	e.isNew = false
	e.isDirty = false
}

// PersistibleActualExpense is a rich domain model for ActualExpense
type PersistibleActualExpense struct {
	ActualExpense
	isNew   bool
	isDirty bool
}

func (e *PersistibleActualExpense) IsNew() bool {
	return e.isNew
}

func (e *PersistibleActualExpense) MarkPersisted() {
	e.isNew = false
	e.isDirty = false
}
