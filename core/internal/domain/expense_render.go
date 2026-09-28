package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/budgets/core/internal/representation"
)

// AmountDecrypter decrypts a persisted encrypted amount into plain Money.
// Implemented by an adapter over the encryption facility; decryption happens
// while constructing the rendering object so Render stays pure.
type AmountDecrypter interface {
	DecryptAmount(ciphertext string) (Money, error)
}

// MoneyConverter converts money between currencies using a historical rate.
// Implemented by the currency marketplace.
type MoneyConverter interface {
	ConvertHistorical(ctx context.Context, amount Money, to Currency, quote QuoteType, date time.Time) (Money, error)
}

// PresentationPrefs captures the display settings used when converting amounts
// for rendering: the user's display currency and preferred quote type.
type PresentationPrefs struct {
	displayCurrency Currency
	preferredQuote  QuoteType
}

// DefaultPresentationPrefs returns the fallback presentation settings used when
// the user has no stored preferences.
func DefaultPresentationPrefs() PresentationPrefs {
	return PresentationPrefs{displayCurrency: CurrencyUSD, preferredQuote: QuoteOfficial}
}

// Convert converts money into the presentation currency at the given date.
// Returns the input money unchanged when no conversion is needed or no
// converter is configured.
func (pp PresentationPrefs) Convert(ctx context.Context, m MoneyConverter, money Money, date time.Time) (Money, error) {
	if m == nil || money.Currency == pp.displayCurrency {
		return money, nil
	}
	return m.ConvertHistorical(ctx, money, pp.displayCurrency, pp.preferredQuote, date)
}

func renderMoney(m Money, converted bool) representation.Money {
	return representation.Money{Amount: m.Amount.String(), Currency: string(m.Currency), Converted: converted}
}

// DecryptedExpectedExpense is a persisted expected expense whose amount has
// been decrypted at construction time. It renders without a converted amount.
type DecryptedExpectedExpense struct {
	expense *PersistedExpectedExpense
	amount  Money
}

func NewDecryptedExpectedExpense(e *PersistedExpectedExpense, d AmountDecrypter) (*DecryptedExpectedExpense, error) {
	amount, err := d.DecryptAmount(e.encryptedAmount)
	if err != nil {
		return nil, err
	}
	return &DecryptedExpectedExpense{expense: e, amount: amount}, nil
}

// Convert returns this expense rendered with its amount converted via the
// presentation preferences at the given as-of date.
func (d *DecryptedExpectedExpense) Convert(ctx context.Context, prefs PresentationPrefs, m MoneyConverter, asOf time.Time) (*ConvertedExpectedExpense, error) {
	converted, err := prefs.Convert(ctx, m, d.amount, asOf)
	if err != nil {
		return nil, err
	}
	return &ConvertedExpectedExpense{decrypted: d, amount: converted}, nil
}

func (d *DecryptedExpectedExpense) Render() Rendered[representation.ExpectedExpense] {
	return Render(representation.ExpectedExpense{
		ID:          d.expense.externalID,
		Name:        d.expense.name,
		Description: d.expense.description,
		Amount:      renderMoney(d.amount, false),
		CategoryID:  d.expense.categoryExternalID,
		CreatedAt:   d.expense.createdAt,
		UpdatedAt:   d.expense.updatedAt,
	})
}

// ConvertedExpectedExpense is a decrypted expected expense rendered with a
// converted amount in the user's display currency.
type ConvertedExpectedExpense struct {
	decrypted *DecryptedExpectedExpense
	amount    Money
}

func (c *ConvertedExpectedExpense) Render() Rendered[representation.ExpectedExpense] {
	v := c.decrypted.Render().Value()
	if c.amount.Currency != c.decrypted.amount.Currency {
		m := renderMoney(c.amount, true)
		v.ConvertedAmount = &m
	}
	return Render(v)
}

// DecryptedActualExpense is a persisted actual expense whose amount has been
// decrypted at construction time.
type DecryptedActualExpense struct {
	expense *PersistedActualExpense
	amount  Money
}

func NewDecryptedActualExpense(e *PersistedActualExpense, d AmountDecrypter) (*DecryptedActualExpense, error) {
	amount, err := d.DecryptAmount(e.encryptedAmount)
	if err != nil {
		return nil, err
	}
	return &DecryptedActualExpense{expense: e, amount: amount}, nil
}

// Convert returns this expense rendered with its amount converted via the
// presentation preferences at the expense's own date.
func (d *DecryptedActualExpense) Convert(ctx context.Context, prefs PresentationPrefs, m MoneyConverter) (*ConvertedActualExpense, error) {
	converted, err := prefs.Convert(ctx, m, d.amount, d.expense.expenseDate)
	if err != nil {
		return nil, err
	}
	return &ConvertedActualExpense{decrypted: d, amount: converted}, nil
}

func (d *DecryptedActualExpense) Render() Rendered[representation.ActualExpense] {
	return Render(representation.ActualExpense{
		ID:          d.expense.externalID,
		Name:        d.expense.name,
		Description: d.expense.description,
		ExpenseDate: d.expense.expenseDate.Format(apiDateFormat),
		Amount:      renderMoney(d.amount, false),
		CategoryID:  d.expense.categoryExternalID,
		CreatedAt:   d.expense.createdAt,
		UpdatedAt:   d.expense.updatedAt,
	})
}

// ConvertedActualExpense is a decrypted actual expense rendered with a
// converted amount in the user's display currency.
type ConvertedActualExpense struct {
	decrypted *DecryptedActualExpense
	amount    Money
}

func (c *ConvertedActualExpense) Render() Rendered[representation.ActualExpense] {
	v := c.decrypted.Render().Value()
	if c.amount.Currency != c.decrypted.amount.Currency {
		m := renderMoney(c.amount, true)
		v.ConvertedAmount = &m
	}
	return Render(v)
}

// BudgetSummary accumulates expected and actual totals for a budget in the
// user's presentation currency. All conversion happens at construction time:
// expected expenses convert at the budget's cutoff date (min(end_date, now)),
// actual expenses at their own expense_date.
type BudgetSummary struct {
	budgetExternalID uuid.UUID
	displayCurrency  Currency
	expectedTotal    decimal.Decimal
	actualTotal      decimal.Decimal
}

func NewBudgetSummary(
	ctx context.Context,
	budget *PersistedBudget,
	expected []*DecryptedExpectedExpense,
	actual []*DecryptedActualExpense,
	prefs PresentationPrefs,
	m MoneyConverter,
) (*BudgetSummary, error) {
	s := &BudgetSummary{
		budgetExternalID: budget.externalID,
		displayCurrency:  prefs.displayCurrency,
		expectedTotal:    decimal.Zero,
		actualTotal:      decimal.Zero,
	}

	asOf := budget.ConversionCutoffAt(time.Now())
	for _, e := range expected {
		converted, err := prefs.Convert(ctx, m, e.amount, asOf)
		if err != nil {
			return nil, err
		}
		s.expectedTotal = s.expectedTotal.Add(converted.Amount)
	}

	for _, e := range actual {
		converted, err := prefs.Convert(ctx, m, e.amount, e.expense.expenseDate)
		if err != nil {
			return nil, err
		}
		s.actualTotal = s.actualTotal.Add(converted.Amount)
	}

	return s, nil
}

func (s *BudgetSummary) Render() Rendered[representation.BudgetSummary] {
	return Render(representation.BudgetSummary{
		BudgetID:      s.budgetExternalID,
		ExpectedTotal: representation.Money{Amount: s.expectedTotal.String(), Currency: string(s.displayCurrency), Converted: true},
		ActualTotal:   representation.Money{Amount: s.actualTotal.String(), Currency: string(s.displayCurrency), Converted: true},
		Difference:    representation.Money{Amount: s.expectedTotal.Sub(s.actualTotal).String(), Currency: string(s.displayCurrency), Converted: true},
	})
}
