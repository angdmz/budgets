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

// NullMoneyConverter is the NullObject MoneyConverter: it returns the money
// unchanged, so pipelines can run without a configured marketplace.
type NullMoneyConverter struct{}

func (NullMoneyConverter) ConvertHistorical(ctx context.Context, amount Money, to Currency, quote QuoteType, date time.Time) (Money, error) {
	return amount, nil
}

// Convert converts money into the presentation currency at the given date.
// Returns the input money unchanged when no conversion is needed.
func (pp PresentationPrefs) Convert(ctx context.Context, m MoneyConverter, money Money, date time.Time) (Money, error) {
	if money.Currency == pp.displayCurrency {
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
	return render(representation.ExpectedExpense{
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
	v := c.decrypted.Render().value()
	if c.amount.Currency != c.decrypted.amount.Currency {
		m := renderMoney(c.amount, true)
		v.ConvertedAmount = &m
	}
	return render(v)
}

// ExpectedExpenseViews is the renderable form of a persisted expected-expense
// list: each item is decrypted and converted at construction time so Render is
// pure. All expenses convert at the same as-of date, so a caching converter
// performs at most one provider call per distinct currency pair.
type ExpectedExpenseViews struct {
	items []*ConvertedExpectedExpense
}

func NewExpectedExpenseViews(ctx context.Context, expenses []PersistedExpectedExpense, d AmountDecrypter, prefs PresentationPrefs, m MoneyConverter, asOf time.Time) (*ExpectedExpenseViews, error) {
	views := &ExpectedExpenseViews{items: make([]*ConvertedExpectedExpense, 0, len(expenses))}
	for i := range expenses {
		decrypted, err := NewDecryptedExpectedExpense(&expenses[i], d)
		if err != nil {
			return nil, err
		}
		converted, err := decrypted.Convert(ctx, prefs, m, asOf)
		if err != nil {
			return nil, err
		}
		views.items = append(views.items, converted)
	}
	return views, nil
}

// ExpectedExpenseViews builds this budget's renderable expected-expense list,
// converting at the budget's conversion cutoff (earlier of end date and now).
func (b *PersistedBudget) ExpectedExpenseViews(ctx context.Context, expenses []PersistedExpectedExpense, d AmountDecrypter, prefs PresentationPrefs, m MoneyConverter, now time.Time) (*ExpectedExpenseViews, error) {
	return NewExpectedExpenseViews(ctx, expenses, d, prefs, m, b.conversionCutoffAt(now))
}

func (v *ExpectedExpenseViews) Render() Rendered[[]representation.ExpectedExpense] {
	out := make([]representation.ExpectedExpense, len(v.items))
	for i, item := range v.items {
		out[i] = item.Render().value()
	}
	return render(out)
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
	return render(representation.ActualExpense{
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
	v := c.decrypted.Render().value()
	if c.amount.Currency != c.decrypted.amount.Currency {
		m := renderMoney(c.amount, true)
		v.ConvertedAmount = &m
	}
	return render(v)
}

// ActualExpenseViews is the renderable form of a persisted actual-expense
// list: each item is decrypted and converted at its own expense_date at
// construction time so Render is pure. A caching converter performs at most
// one provider call per distinct (currency, quote, date) tuple.
type ActualExpenseViews struct {
	items []*ConvertedActualExpense
}

func NewActualExpenseViews(ctx context.Context, expenses []PersistedActualExpense, d AmountDecrypter, prefs PresentationPrefs, m MoneyConverter) (*ActualExpenseViews, error) {
	views := &ActualExpenseViews{items: make([]*ConvertedActualExpense, 0, len(expenses))}
	for i := range expenses {
		decrypted, err := NewDecryptedActualExpense(&expenses[i], d)
		if err != nil {
			return nil, err
		}
		converted, err := decrypted.Convert(ctx, prefs, m)
		if err != nil {
			return nil, err
		}
		views.items = append(views.items, converted)
	}
	return views, nil
}

func (v *ActualExpenseViews) Render() Rendered[[]representation.ActualExpense] {
	out := make([]representation.ActualExpense, len(v.items))
	for i, item := range v.items {
		out[i] = item.Render().value()
	}
	return render(out)
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
	now time.Time,
) (*BudgetSummary, error) {
	s := &BudgetSummary{
		budgetExternalID: budget.externalID,
		displayCurrency:  prefs.displayCurrency,
		expectedTotal:    decimal.Zero,
		actualTotal:      decimal.Zero,
	}

	asOf := budget.conversionCutoffAt(now)
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
	return render(representation.BudgetSummary{
		BudgetID:      s.budgetExternalID,
		ExpectedTotal: representation.Money{Amount: s.expectedTotal.String(), Currency: string(s.displayCurrency), Converted: true},
		ActualTotal:   representation.Money{Amount: s.actualTotal.String(), Currency: string(s.displayCurrency), Converted: true},
		Difference:    representation.Money{Amount: s.expectedTotal.Sub(s.actualTotal).String(), Currency: string(s.displayCurrency), Converted: true},
	})
}
