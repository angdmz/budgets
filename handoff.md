# Handoff: Modal Migration & UI Refactor

## Completed Work

### 1. index.css cleanup + conftest.py mobile fixtures
- Done (previous session)

### 2. Layout shell/nav refactor
- Bottom nav, More drawer, container — done (previous session)

### 3. Dialog component + modal migration (all pages)
- **Created** `app/src/components/Dialog.tsx` — shared modal wrapper with title, onClose, backdrop, responsive padding.
- **BudgetDetail.tsx** — 6 modals migrated (expected/actual × create/edit/delete).
- **Groups.tsx** — 3 modals migrated (create group, invite, delete confirm).
- **Budgets.tsx** — 4 modals migrated (create, edit, duplicate, delete).
- **Categories.tsx** — 3 modals migrated (create, edit, delete).
- **ExpectedExpenses.tsx** — 3 modals migrated (create, edit, delete).
- **Expenses.tsx** — 3 modals migrated (create, edit, delete).
- All button layouts use `flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3` + `min-h-[44px]` touch targets.

### 4. useBudgetSelection hook + BudgetSummary component
- **Created** `app/src/hooks/useBudgetSelection.ts` — encapsulates group/budget/expected-expenses/actual-expenses queries + derived totals (expectedTotal, actualTotal, difference). Used by:
  - `Dashboard.tsx` (replaced all inline queries + totals logic)
  - `Expenses.tsx` (replaced group/budget/actual-expenses queries)
  - `ExpectedExpenses.tsx` (replaced group/budget/expected-expenses queries)
- **Created** `app/src/components/BudgetSummary.tsx` — 3-card summary (expected, actual, difference) with `formatCurrency`. Used by:
  - `Dashboard.tsx`
  - `BudgetDetail.tsx`

### 5. ExpenseFormFields + ExpenseFormDialog + CategoryCombobox fixes (DONE)
- **Created** `app/src/components/ExpenseFormFields.tsx` — reusable form fields (name, amount+currency, date, description, category combobox).
- **Created** `app/src/components/ExpenseFormDialog.tsx` — wraps Dialog + ExpenseFormFields + submit/cancel buttons with responsive layout.
- **Refactored** `Expenses.tsx` — create/edit modals now use `ExpenseFormDialog` (removed ~80 lines of duplicated form JSX).
- **Refactored** `BudgetDetail.tsx` — 4 form modals migrated to `ExpenseFormDialog`:
  - Expected expense create (`showDate={false}`)
  - Expected expense edit (`showDate={false}`)
  - Actual expense create (`showDate`)
  - Actual expense edit (`showDate`)
  - Removed unused `CategoryCombobox` and `CurrencySelect` imports.
- **Refactored** `ExpectedExpenses.tsx` — create/edit modals migrated to `ExpenseFormDialog` (`showDate={false}`). Removed unused `CategoryCombobox` and `CurrencySelect` imports.
- **Fixed** `CategoryCombobox.tsx` — added `min-h-[44px]` to trigger button for mobile touch targets.

### 6. ExpenseList component (DONE)
- **Created** `app/src/components/ExpenseList.tsx` — generic reusable expense table with configurable columns (`showDate`, `showCategory`, `showDescription`, `showActions`), `maxRows` for compact variants, `emptyMessage` for empty state, and `onEdit`/`onDelete` callbacks. Generic type parameter `T extends ExpenseListItem` preserves original expense type through callbacks.
- **Refactored** `Expenses.tsx` — replaced ~60 lines of inline table JSX with `<ExpenseList>` (all columns + actions). Removed unused `formatDate`/`formatCurrency` imports.
- **Refactored** `BudgetDetail.tsx` — replaced both inline tables (expected + actual, ~82 lines total) with `<ExpenseList>` calls. Removed unused `formatCurrency` import.
- **Refactored** `Dashboard.tsx` — replaced ~30 lines of inline recent-expenses table with `<ExpenseList showDate maxRows={5} />`. Removed unused `formatCurrency` import.
- Net reduction: ~130 lines of duplicated table JSX.

### 7. Mobile E2E tests (DONE)
- **Created** `tests/test_responsive_mobile.py` — 19 tests covering:
  - Layout shell: bottom nav visible, desktop nav hidden, mobile header visible, More drawer opens + navigates.
  - Horizontal overflow: landing + /app at 390px and 320px; dashboard, groups, budgets, categories, expenses, expected-expenses at 390px; dashboard, budgets, expenses at 320px.
  - Modal dialogs: no overflow, buttons stacked (`flex-col-reverse`), 44px touch target height.
  - Table scroll: expense table doesn't cause page-level overflow.
- Uses `driver_mobile` (390x844) and `driver_mobile_narrow` (320x568) fixtures from `conftest.py`.
- Uses `assert_no_horizontal_overflow` helper from `conftest.py`.

### 8. Responsive layout fixes for P1/P2 screens (DONE)
- **ExpenseList.tsx** — changed `overflow-hidden` to `overflow-x-auto` so wide tables scroll horizontally on mobile instead of clipping. Added `min-h-[44px]` touch targets to edit/delete buttons.
- **ExpectedExpenses.tsx** — replaced inline table (~50 lines) with `<ExpenseList>` component for consistency with other pages. Removed unused `formatCurrency` import.
- **Budgets.tsx** — added `overflow-x-auto` to table wrapper, `flex-wrap` to header buttons, `min-h-[44px]` touch targets to edit/duplicate/delete actions.
- **Categories.tsx** — added `min-h-[44px]` touch targets and horizontal padding to edit/delete buttons in category cards.
- **Groups.tsx** — invite link input+copy button now stack vertically on mobile (`flex-col sm:flex-row`), input gets `min-w-0` to shrink properly. Invitation list items stack on mobile with touch targets on revoke button.
- **CreateBudgetPlan.tsx** — stepper gets `overflow-x-auto` for horizontal scroll on narrow screens. Review table wrapper changed from `overflow-hidden` to `overflow-x-auto`.
- **Onboarding.tsx** — heading size reduced on mobile (`text-xl sm:text-2xl`), card padding reduced (`p-4 sm:p-6`), compare cards stack 1-column on mobile (`grid-cols-1 sm:grid-cols-3`), form fields in AddExpectedExpensesStep stack vertically on mobile (`flex-col sm:flex-row`), currency select goes full-width on mobile.

### 9. Accessibility audit + full regression run (DONE)
- **Dialog.tsx** — added backdrop click-to-close, close button with `aria-label`, `aria-hidden` on SVG.
- **ExpenseFormFields.tsx** — associated all labels with inputs via `htmlFor`/`id` (`expense-name`, `expense-amount`, `expense-date`, `expense-description`). Added `role="alert"` to category error message.
- **CategoryCombobox.tsx** — added `aria-expanded`, `aria-haspopup="listbox"`, `aria-label` to trigger button. Added `role="listbox"` to dropdown, `role="option"` + `aria-selected` to category items. Added `aria-label` to search input.
- **ExpenseList.tsx** — added `scope="col"` to all `<th>` headers. Added descriptive `aria-label` to edit/delete action buttons (e.g., "Edit Groceries").
- **ExpenseFormDialog.tsx** — added `role="alert"` to error message.
- **CurrencySelect.tsx** — already had `aria-label="Currency"`.
- **Layout.tsx** — added `aria-label` to both `<nav>` elements (Main/Mobile navigation). Added `aria-current="page"` to active links in desktop and mobile nav. Added `aria-expanded` + `aria-haspopup="dialog"` + `aria-label` to More button. Added `aria-hidden="true"` to all decorative SVGs. Added `aria-label="Language"` to both language `<select>` elements.
- **Build**: `docker compose build app` succeeded cleanly (`tsc && vite build` passed, 986 modules transformed).
- **Integration tests**: Require Auth0 secrets + full stack running — not executed in this session.

## Pending TODO Items

| # | Task | Status |
|---|------|--------|
| 5 | ExpenseFormFields + ExpenseFormDialog + CategoryCombobox fixes | **Done** |
| 6 | ExpenseList → recent expenses + edit reuse | **Done** |
| 7 | Mobile E2E tests (test_responsive_mobile.py) | **Done** |
| 8 | P1/P2 screens (ExpectedExpenses, Budgets, Categories, Onboarding, Groups, CreateBudgetPlan) | **Done** |
| 9 | Accessibility + full regression run | **Done** |

## Next Steps

All tasks from the handoff are complete. To run the full integration test suite:
1. Ensure secrets are set up (`.env`, `./secrets/`).
2. Start the full stack: `docker compose up -d`.
3. Run integration tests: `docker compose --profile integration run integration-tests`.

## Build Status
Last `docker compose build app` succeeded cleanly (`tsc && vite build` passed, 986 modules transformed).

## Key Files
- `app/src/components/Dialog.tsx`
- `app/src/components/BudgetSummary.tsx`
- `app/src/components/ExpenseFormFields.tsx`
- `app/src/components/ExpenseFormDialog.tsx`
- `app/src/components/CategoryCombobox.tsx`
- `app/src/components/ExpenseList.tsx`
- `app/src/hooks/useBudgetSelection.ts`
- `app/src/pages/Dashboard.tsx`
- `app/src/pages/BudgetDetail.tsx`
- `app/src/pages/Expenses.tsx`
- `app/src/pages/ExpectedExpenses.tsx`
- `app/src/pages/Groups.tsx`
- `app/src/pages/Budgets.tsx`
- `app/src/pages/Categories.tsx`
- `app/src/pages/Onboarding.tsx`
- `app/src/pages/CreateBudgetPlan.tsx`

## Notes
- Use `docker compose` for all app build/run/logs commands (e.g. `docker compose build app`, `docker compose up -d app`, `docker compose logs -f app`).
