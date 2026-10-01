# Getter removal: problem, reasoning, solution, and implementation plan

Status: **implementation complete**. All getter-bearing `Persisted*` types were migrated to the
`Rendered[T]`/`representation` rendering model; the full test suite passes
(`docker compose --profile test run --rm test`: 152 tests, 30 skipped, 0 failures). Phases 4
(documents/jobs/drafts/worker) and parts of the plan that referenced `document.go` were N/A — those
types do not exist in this codebase. Historical detail below is kept for reference.

**Rev 2 (identity design changed):** the `Ref[T]`/`ExternalRef[T]`/`FKBinder` abstraction was dropped.
All ID management is delegated to the `Persistible*`/`Persisted*` objects themselves — domain cores
carry no identity at all (see §3.3). Rendering (`Rendered[T]`, `representation` package, decorator
chains) is unchanged.

Companion doc for the currency feature: `handoff.md` (repo root).

---

## 1. The problem

`core/AGENTS.md` contradicts itself:

- **Line 6** forbids getters outright ("instead of using getters, we should use methods to modify
  the objects, and never expose the struct fields...").
- **Lines 26-27** allow getters on `Persisted*` for business data and `externalID`.
- **Line 39** instructs handlers to build API responses from getter methods.

The code follows lines 26-27/39. Getters are everywhere. The goal is to **remove all getters**
while keeping tell-don't-ask: objects act on their own state; callers never read state to decide
or to serialize.

### 1.1 Where the contradiction manifests in code

Getter usage falls into three groups:

**Group A — getters used only to build API responses** (follows 26/27/39, breaks the wording of 6):

Getter definitions:
- `core/internal/domain/persistible.go:451-469` — `PersistedGroup` (`ExternalID`, `Name`, `Description`, `CreatedAt`, `UpdatedAt`)
- `core/internal/domain/persistible.go:528-554` — `PersistedCategory` (+ `Color`, `Icon`)
- `core/internal/domain/persistible.go:621-647` — `PersistedBudget` (+ `StartDate`, `EndDate`)
- `core/internal/domain/persistible.go:718-744` — `PersistedExpectedExpense` (+ `EncryptedAmount`, `CategoryExternalID`)
- `core/internal/domain/persistible.go:826-856` — `PersistedActualExpense` (+ `ExpenseDate`)
- `core/internal/domain/persistible.go:1310-1336` — `PersistedUserPreference` (`Theme`, `Language`, `DisplayCurrency`, `PreferredQuoteType`, timestamps)
- `core/internal/domain/invitation.go:248-290` — `PersistedInvitation` (`Token`, `GroupName`, `InviterName`, `Status`, `Role`, `ExpiresAt`, `AcceptedAt`, timestamps)
- `core/internal/domain/document.go:226-237` — `PersistedExpenseDocument` (`BudgetExternalID`, `UploadedByUserID`, `OriginalFilename`, `ContentType`, `FileSize`, `DocumentType`, `EncryptedStoragePath`, `Status`, `ErrorMessage`, timestamps)
- `core/internal/domain/document.go:340-346` — `PersistedProcessingJob` (`DocumentID`, `JobType`, `Status`, `Attempts`, `MaxAttempts`, `ErrorMessage`)
- `core/internal/domain/document.go:431-440` — `PersistedExpenseDraft` (`DocumentExternalID`, `BudgetExternalID`, `CategoryExternalID`, `Status`, `ExtractionMethod`, `Confidence`, `ConfirmedActualExpenseID`, timestamps)
- `core/internal/domain/exchange_rate.go:130-139` — `PersistedExchangeRate` (`ID`, `FromCurrency`, `ToCurrency`, `Quote`, `Rate`, `Provider`, `ObservedAt`, timestamps)
- `core/internal/domain/onboarding.go:161-166` — `PersistedUserOnboarding` (`ID`, `UserID`, `Status`, `CurrentStep`, timestamps)
- `core/internal/domain/onboarding.go:377-380` — `PersistedUserOnboardingStep` (`ID`, `OnboardingID`, `Step`, `Status`, `CreatedAt`)

Handlers building responses from them:
- `core/internal/handler/expense_helpers.go:46-73` — `toExpectedExpenseResponse`, `toActualExpenseResponse`
- `core/internal/handler/category_handler.go:144-155`, `:231-239`
- `core/internal/handler/budget_handler.go:229-237` (and likely create/update/get flows nearby)
- `core/internal/handler/invitation_handler.go:72-82`, `:140-151`, `:249-255`
- `core/internal/handler/document_handler.go:143-148`, `:200-206`, `:259-265`, `:322-328`
- `core/internal/handler/onboarding_handler.go:274-289`

**Group B — getters used to make decisions or extract state for side effects** (violates the intent of
line 6 — "ask" then act):

- `core/internal/handler/invitation_handler.go:245`: `invitation.IsExpired() || invitation.Status() == domain.InvitationStatusRevoked` → domain must answer this itself, e.g. `invitation.EnsureUsable(now)`.
- `core/internal/handler/document_handler.go:383`: `storagePath = doc.EncryptedStoragePath()` — handler extracts the path to delete the file after the tx.
- `expense.EncryptedAmount()` pulled to decrypt in handlers at `expense_handler.go:94,150,245,343,474,530,615,719,889,905` (plus the same pattern in `expense_helpers.go` callers).
- `expense_handler.go:843-844`: `pref.DisplayCurrency()`, `pref.PreferredQuoteType()` read to drive conversion; `:854`: `budget.EndDate()`; `:911`: `e.ExpenseDate()` read inside a conversion loop.
- `core/cmd/worker/main.go:88-90`: `job.ExternalID()`, `job.JobType()`, `job.DocumentID()`, `job.Attempts()` read to log and dispatch `processJob`.

**Group C — getters that expose internal `int64` IDs** (breaks AGENTS.md lines 22 and 26 as well as 6 —
"internal `id` is a local variable, never exposed"):

- `PersistedUserOnboarding.ID()` / `.UserID()` (`onboarding.go:161,163`) — used at
  `onboarding_handler.go:110` and `:177` as `NewPersistibleUserOnboardingStep(ob.ID(), step, data)` /
  `NewSkippedUserOnboardingStep(ob.ID(), step)` to wire the FK.
- `PersistedUserOnboardingStep.ID()` / `.OnboardingID()` (`onboarding.go:377,379`).
- `PersistedInvitation.GroupID()` (`invitation.go:252`) — used at `invitation_handler.go:197` where a
  handler runs raw SQL `SELECT external_id FROM budgeting_groups WHERE id = $1` with it (also breaks
  "domain objects own their SQL", line 20).
- `PersistedExchangeRate.ID()` (`exchange_rate.go:130`).
- `PersistedExpenseDocument.UploadedByUserID()` (`document.go:228`).
- `PersistedProcessingJob.DocumentID()` (`document.go:341`) — used by `worker/main.go:90`.
- `PersistedExpenseDraft.ConfirmedActualExpenseID()` (`document.go:438`) — checked for use in
  `document_handler.go` confirm flow; `MarkConfirmed` already exists at `document.go:451-456`.

Also relevant: `PersistedUserPreference` holds `userID int64` as a struct field (`persistible.go:1285`),
which technically violates "domain objects do NOT hold database IDs" (line 22) — it stores the *parent's*
internal ID rather than its own, presumably needed by `UpdateIn`. This needs a dedicated decision (see §6.5).

### 1.2 Current response types (destination for rendered output)

All API response structs live in `core/internal/handler/dto.go` (e.g. `GroupResponse:24`,
`CategoryResponse:46`, `BudgetResponse:70`, `BudgetSummaryResponse:80`, `MoneyResponse:92`,
`ExpectedExpenseResponse:112`, `ActualExpenseResponse:141`, `PreferenceResponse:171`,
`InvitationResponse:205`, `InvitationDetailResponse:218`, `OnboardingResponse:235`,
`OnboardingStepResponse:244`) and `core/internal/handler/document_handler.go` (`DocumentResponse:410`,
`DraftResponse:423`). Swagger `@Success` annotations reference these types.

Repository-style code remains in `core/internal/repository/` (`user.go`, `interfaces.go`,
`pagination.go`) — legacy per AGENTS.md line 34; the migration only covers `Persisted*` types.

---

## 2. Reasoning behind the chosen design

Getters exist for three needs: (a) identity (internal IDs, FKs, external UUIDs), (b) output
(serialization), (c) decisions/side effects. Each needs its own replacement:

- **Identity → persisted objects own all IDs.** (REVISED in Rev 2 — the earlier `Ref[T]`/
  `ExternalRef[T]`/`FKBinder` design was dropped.) There is no generic reference abstraction. Domain
  cores carry no identity; only `Persistible*`/`Persisted*` structs hold `id`/`externalID` as
  unexported fields and bind them into their own SQL. FK type-safety comes from unexported named ID
  types (`type budgetID int64` — pgx v5 encodes named int64 kinds directly, no `driver.Valuer`
  needed). FKs known only by UUID bind as inline subqueries in the object's own SQL, removing the
  extra `SELECT id ...` lookup (helps the O(1) I/O rule). Parents create children in-package
  (`ob.NewStep(...)` reads `ob.id` — same package, not a getter).
- **Output → specialized render objects.** Decided in conversation (supersedes earlier "exporter sink"
  and "post-render `Extended` wrapper" ideas): a rendering is **final**; handlers never read, modify,
  or extend it. When more data must appear in output, a *more specific domain object* takes the extra
  data as collaborators and renders itself. This keeps tell-don't-ask airtight (the handler literally
  cannot observe object state) and keeps all fallible work/I/O at construction, leaving `Render()`
  pure.
- **Decisions → intention-revealing methods.** `EnsureUsable`, `RunWith`, `DeleteFrom(..., trash)`,
  `parent.NewChild(...)` replace `Status() ==`, path extraction, and FK lookups.
- **Alternative rejected: Identity Map** (Fowler). `map[any]identity` keyed by object pointer —
  hidden state, fragile pointer keys, still requires asking the map. `Ref[T]` keeps identity
  encapsulated inside the object with less machinery.
- **Rejected variant: `MarshalJSON` on domain objects.** Swagger can't inspect the shape; JSON tags
  on domain structs were explicitly avoided (line 39 context). Hence a separate `representation`
  package of pure wire shapes.

Key invariants:

1. **Renderings are final.** No post-render extension mechanism (`Extended`/byte-splice) exists.
   Request-scoped extras (links, per-user permissions) go through the same collaborator mechanism:
   `budget.WithLinks(linker)`.
2. **`Render()` cannot fail and does no I/O.** All fallible work happens when the specific object is
   constructed (`DecryptWith`, `ConvertWith`). This also keeps O(1) I/O (line 13): rate lookups batch
   before construction, never per-render.
3. **Internal IDs never leave `domain`.** `id`/`externalID` are unexported fields on `Persisted*`
   structs only; they appear in SQL args and scan targets but are never a Go-visible value outside
   the package and are never returned by any method.
4. **No state-splitting flags** (line 23 + AGENTS.md's granular-objects/polymorphism rule):
   lifecycle stages are separate structs (`Persistible*`/`Persisted*`); status-based decisions are
   separate types/methods (`EnsureUsable`, `MarkFailed`), not `if x.Status() ==` checks at call
   sites.

---

## 3. The explicit solution

### 3.1 `Rendered[T]` — opaque output value

```go
// core/internal/domain/rendered.go
package domain

import "encoding/json"

// Rendered is an opaque, final rendering. Only MarshalJSON escapes.
type Rendered[T any] struct{ v T }

func (r Rendered[T]) MarshalJSON() ([]byte, error) { return json.Marshal(r.v) }
func (r Rendered[T]) Value() T                      { return r.v } // internal use (tests, nested composition) — REVISIT: keep unexported-eyes-only by convention, or make unexported field accessor method for package-internal consumers. Decide at implementation time.
```

- Public render entry point per object: `func (b *PersistedBudget) Render() Rendered[representation.Budget]`.
- Composition inside domain uses unexported `render() T` methods returning the raw shape; parents embed
  child shapes (they're in-package, no getters needed).
- Handlers do `c.JSON(http.StatusOK, response)` — `Rendered` marshals to the same JSON as before.

### 3.2 `representation` package — wire shapes

New package `core/internal/representation/` containing pure data structs mirroring today's
`*Response` DTOs, with JSON tags, no domain deps:

```go
package representation

import (
	"time"
	"github.com/google/uuid"
)

type Money struct {
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
	Converted bool   `json:"converted,omitempty"`
}

type ActualExpense struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	ExpenseDate string    `json:"expense_date"`
	Amount      Money     `json:"amount"`
	CategoryID  uuid.UUID `json:"category_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Extension via plain struct embedding (NO Marshaler on inner types — avoids the
// encoding/json promotion pitfall where an embedded Marshaler suppresses sibling fields;
// only the outer Rendered has MarshalJSON).
type ConvertedActualExpense struct {
	ActualExpense
	ConvertedAmount Money `json:"converted_amount"`
}
```

Plan for DTOs: **move** `handler/dto.go` response structs into `representation` (keep request DTOs and
`ErrorResponse` in handler), update swagger annotations to `representation.X`, or leave annotations
pointing at thin aliases (`type BudgetResponse = representation.Budget`) if that churns less. Since
`Rendered[T]` marshals identically, existing API-level tests assert the same JSON and serve as the
safety net.

### 3.3 ID ownership: persisted/persistible objects only — NO `Ref[T]`

Decision (Rev 2; supersedes the `Ref[T]`/`ExternalRef[T]`/`FKBinder` design): there is **no generic
reference abstraction**. All ID management is delegated to the `Persistible*`/`Persisted*` objects;
the domain core carries no identity at all.

```go
// pure domain core: no identity, no SQL — validation and business behavior live here
type budget struct {
	name        string
	description string
	startDate   time.Time
	endDate     time.Time
}

func (b *budget) UpdateName(n string)           { b.name = n }
func (b *budget) render() representation.Budget { /* ID field left empty */ }

// creation stage: core + insert ability; still no IDs
type PersistibleBudget struct{ budget }
func (b *PersistibleBudget) PersistTo(ctx context.Context, p Persister) (*PersistedBudget, error)

// loaded stage: core + identity; owns every ID-related operation
type PersistedBudget struct {
	budget
	id                   int64
	externalID           uuid.UUID
	createdAt, updatedAt time.Time
}

func (b *PersistedBudget) Render() Rendered[representation.Budget] {
	r := b.budget.render()
	r.ID = b.externalID // persisted stage adds its own identity to the output
	return Rendered[representation.Budget]{v: r}
}
```

**FK type-safety without generics:** unexported named ID types:

```go
type budgetID   int64
type categoryID int64
type expenseID  int64
```

pgx v5 encodes/scans named types whose underlying kind is `int64` — no `driver.Valuer` needed.
Passing a `categoryID` where a `budgetID` is expected is a compile error.

**FKs known only by UUID** (e.g. `category_id` from a request body): the persistible/persisted object
keeps `categoryExternalID uuid.UUID` and binds an inline subquery in its own SQL:

```sql
INSERT INTO actual_expenses (..., category_id, ...)
VALUES (..., (SELECT id FROM expense_categories WHERE external_id = $5 AND revoked_at IS NULL), ...)
```

This replaces the separate `SELECT id ...` round trip in today's `UpdateIn`
(`persistible.go:763-772`, `:879-888`) — an O(1) I/O win. 0 rows → `ErrNotFound`.

**Parent → child wiring:** same-package field access, not a getter. `ob.NewStep(step, data)` reads
`ob.id` directly to seed the child's FK. Go encapsulates per-package; every `Persisted*` type lives
in `domain`.

**Rules this creates:**
- All `Persistible*`/`Persisted*` types must live in package `domain` (in-package field access is
  the FK mechanism).
- No method on any `Persisted*` type returns raw `int64`/`uuid.UUID` IDs — external IDs escape only
  inside rendered output. Enforce with a CI grep, e.g.
  `grep -rn 'func (.* \*Persisted\w*) \w*(.*) (int64|uuid.UUID)' internal/domain`.
- The embedded core exposes only an **unexported** `render()`; public `Render()` lives on
  `Persisted*` and merges in identity — otherwise Go embedding promotes a core `Render()` that
  emits output without an ID.

### 3.4 Decorator chain — "more specific object with collaborators"

`Persisted*` is itself the first "more specific object" (core + identity). Every further level of
output is another domain object wrapping the previous one, holding collaborator-produced state,
constructed via a fallible method; `render()` is pure:

```go
// domain collaborator interfaces (impls live in handler/encryption/exchange packages)
type AmountDecrypter interface { Decrypt(encrypted string) (Money, error) }
type RateTable interface {
	Convert(m Money, to Currency, at time.Time) (Money, error) // impl may batch-resolve rates
}

type DecryptedActualExpense struct {
	expense *PersistedActualExpense
	money   Money
}

func (e *PersistedActualExpense) DecryptWith(d AmountDecrypter) (DecryptedActualExpense, error) {
	m, err := d.Decrypt(e.encryptedAmount)
	if err != nil { return DecryptedActualExpense{}, err }
	return DecryptedActualExpense{expense: e, money: m}, nil
}

func (e DecryptedActualExpense) render() representation.ActualExpense { /* fills Amount from e.money */ }
func (e DecryptedActualExpense) Render() Rendered[representation.ActualExpense] { ... }

type ConvertedActualExpense struct {
	base      DecryptedActualExpense
	converted Money
}

func (e DecryptedActualExpense) ConvertWith(rates RateTable, pref PresentationPrefs) (ConvertedActualExpense, error) { ... }
// PresentationPrefs: collaborator interface over PersistedUserPreference so display currency/quote
// are "told" into conversion rather than read via pref.DisplayCurrency()/PreferredQuoteType().

func (c ConvertedActualExpense) render() representation.ConvertedActualExpense {
	return representation.ConvertedActualExpense{ActualExpense: c.base.render(), ConvertedAmount: c.converted.render()}
}
```

Handler usage:

```go
decrypted, err := expense.DecryptWith(h.decrypter)     // decrypter adapts h.encryptor
if err != nil { return err }
converted, err := decrypted.ConvertWith(h.rates, pref) // or skips conversion when not needed
if err != nil { return err }
response = converted.Render()
```

Money itself: `domain.NewMoney` already exists (`expense_handler.go:894` uses it); give Money a
`render() representation.Money` too.

**Combination management:** while combinations are few, concrete chains are fine. If they grow,
compose decorators over an unexported `renderable[T] interface { render() T }` so each decorator wraps
any base.

### 3.5 Composites

Parents compose children inside-domain via `render()`:

```go
type PersistedBudgetWithExpenses struct {
	budget   *PersistedBudget
	actual   []PersistedActualExpense
	expected []PersistedExpectedExpense
}

func (bw *PersistedBudgetWithExpenses) render(d AmountDecrypter, rates RateTable) (representation.BudgetDetail, error) { ... }
```

If different levels need different collaborators, construct children decorators first
(`[]DecryptedActualExpense` → `[]ConvertedActualExpense` → composite render) — all fallible work stays
in the constructor stage.

### 3.6 Replacing decision/side-effect getters

| Today (ask) | Replacement (tell) |
|---|---|
| `invitation.IsExpired() \|\| invitation.Status() == Revoked` → `ErrGone` (`invitation_handler.go:245`) | `if err := invitation.EnsureUsable(now); err != nil` |
| `storagePath = doc.EncryptedStoragePath(); doc.DeleteFrom(ctx,p)` (`document_handler.go:383`) | `doc.DeleteFrom(ctx, p, trash)` where `trash` is a `FileTrash` collector (domain iface); handler calls `trash.EmptyInto(storage)` post-commit |
| `NewPersistibleUserOnboardingStep(ob.ID(), step, data)` (`onboarding_handler.go:110,177`) | `ob.NewStep(step, data)` / `ob.NewSkippedStep(step)` — parent passes its own unexported `id` into the child's FK (same package); no `ID()` getter |
| `processJob(ctx, pool, job.DocumentID(), job.JobType())` (`worker/main.go:90`) | `job.RunWith(ctx, p, processor)` — `processor` is a domain `JobProcessor` iface; job passes its unexported `documentID`/`documentExternalID` internally |
| `job.Attempts()`/`MaxAttempts()` branching | `job.MarkFailed(ctx, p, err)` decides retry vs dead-letter internally |
| `invitation.GroupID()` + raw SQL (`invitation_handler.go:194-198`) | invitation SELECT joins group and scans `g.external_id` into `groupExternalID`; rendering includes group UUID; handler SQL deleted |
| `pref.DisplayCurrency()`, `pref.PreferredQuoteType()` | pass `pref` as `PresentationPrefs` collaborator into converters/renderers |
| `e.EncryptedAmount()` → `DecryptMoney` | `DecryptWith` decorator chain (§3.4) |
| `draft.ConfirmedActualExpenseID()` | if needed for response, draft's `render()` joins/render external ID; internal id never returned |

### 3.7 Batch/N+1 care

- List endpoints: build collaborators once (`RateTable` fetched for all `(currency,date)` pairs or via
  the MultiProvider per currency) then map expenses through `DecryptWith`/`ConvertWith`. O(1) I/O rule
  applies to collaborator construction too.
- `RateTable` may itself encapsulate a lazy-cache with a `Prefetch(ctx, keys)` method called once.
- Keep the existing per-item decrypt loop acceptable ONLY if `Decrypt` is non-I/O (it is — local
  encryption transform); rate conversion IS I/O → must batch or use cached rates.

### 3.8 New AGENTS.md wording (apply at the END of migration)

```markdown
 - we favor tell don't ask: domain objects expose no getters and no exported fields; they render themselves via `Render()` into an opaque `Rendered[T]` whose shape `T` lives in the `representation` package (json tags only there)
 - renderings are final: handlers never read, modify or extend a rendering; when more data is needed, build a more specific domain object (e.g. `expense.DecryptWith(d).ConvertWith(rates, pref)`) that receives the extra data as collaborators and renders itself
 - all fallible work and I/O happen when the specific object is constructed; `Render()` is pure and cannot fail
 - identity is a persistence-stage concern: domain cores have no IDs; only `Persistible*`/`Persisted*` structs hold `id`/`externalID` as unexported fields (prefer unexported named ID types like `type budgetID int64`), bind them into their own SQL, and render external IDs; internal int64 IDs are never returned by any method and never leave the domain package
 - FKs known only by external UUID bind as inline subqueries in the object's own SQL; parent→child FK wiring happens via in-package field access (`ob.NewStep(...)`) — no generic Ref abstraction
 - domain-level decisions stay in domain methods (`EnsureUsable`, `IsUsable`); callers must never branch on object state
```

Update lines 26-27 and 39 to match (26-27 describe persisted-owned identity + no getters; 39 becomes
"handlers pass `Rendered[T]` straight to `c.JSON`; response shapes live in `representation`"). Keep
the existing granular-objects/polymorphism bullet unchanged — it already covers splitting
status-based branches into types.

---

## 4. Verification strategy

- **Safety net = existing API tests** (`handler/*_test.go`, `internal/integration/`): they assert JSON;
  `Rendered[T]` produces identical JSON so they should pass unchanged. Run them after each phase.
- Unit tests for `Rendered[T].MarshalJSON`, inline-FK-subquery SQL fragments (0-rows → `ErrNotFound`
  mapping), named ID type usage, and decorator chain construction errors.
- `go build ./...` + `go vet` + `go test ./...` from `core/` after each phase.
- Optionally run `swag init` (if used in CI) to confirm annotations still resolve — check
  `core/docs`/`docs.go` generation command in the repo before assuming.

---

## 5. Implementation plan (phases, each independently compilable)

Work in `core/`. Prefer one domain type per phase; do NOT interleave phases — each ends with
`go build ./... && go test ./...` green.

**Phase 0 — skeleton (no behavior change)**
1. Create `core/internal/representation/` package; move response DTOs from `handler/dto.go` +
   `document_handler.go` response structs (keep `ErrorResponse`, request DTOs in handler; handler may
   `type X = representation.X` alias to avoid sweeping renames if desired — decide and record).
2. Create `core/internal/domain/rendered.go` with `Rendered[T]`.
3. NO `ref.go`. Optionally create `core/internal/domain/ids.go` with the unexported named ID types
   (`type budgetID int64` etc.); only where an aggregate benefits. No generic ID abstraction.

**Phase 1 — pilot: `PersistedBudget` + `PersistedCategory`** (no collaborators needed)
- Extract the pure core structs (`budget`, `category`), embed into `Persisted*`; add unexported
  `render()` on the core + public `Render()` on `Persisted*`; migrate `budget_handler.go`,
  `category_handler.go`; delete their getters.

**Phase 2 — expenses** (decorator chain)
- `AmountDecrypter` iface; adapter over `h.encryptor`; `DecryptedActualExpense`,
  `ConvertedActualExpense` (+ expected-expense equivalents); migrate `expense_handler.go` CRUD +
  `expense_helpers.go` (delete `toXResponse` helpers); delete expense getters.

**Phase 3 — summary endpoint**
- `BudgetSummary`-rendering domain object taking `(expenses, decrypter, rates, pref)`; move the
  `expense_handler.go:887-925` loop into domain (`TotalAccumulator`/`Summary`); eliminates
  `DisplayCurrency`/`PreferredQuoteType`/`EndDate`/`ExpenseDate` reads from the handler.

**Phase 4 — documents/jobs/drafts + worker**
- `FileTrash`, `JobProcessor` ifaces; `doc.DeleteFrom(..., trash)`; `job.RunWith`; draft render incl.
  external IDs for document/budget/category; delete those getters.

**Phase 5 — onboarding**
- `ob.NewStep`, `ob.NewSkippedStep`, onboarding `Render`; delete `ID()`/`UserID()`/`OnboardingID()`.

**Phase 6 — invitation**
- `EnsureUsable`; join-scan `g.external_id` into `groupExternalID`; delete `GroupID()` + handler SQL.

**Phase 7 — preference + exchange rate + stragglers**
- `PresentationPrefs` collaborator; `PersistedExchangeRate` render if it ever needs output; sweep
  remaining getters via `grep -rn 'func (.* Persisted\w*) \w*()' --include='*.go'` pattern and
  `grep -rn '\.\(Name\|ExternalID\|Status\|ID\|Token\|...\)( )'` call-site search.

**Phase 8 — AGENTS.md update** (§3.8 wording) — DONE ahead of the code migration (`core/AGENTS.md`
lines 6-10, 31-34, 45 now state the target rules). Final `docker compose` build + test run is green.

---

## 6. Open decisions to resolve at implementation time

1. `Rendered[T]` accessibility: keep `v` fully unexported and add a package-visible `peek() T`? Domain
   tests currently construct `PersistedUserPreference` literals directly (see
   `domain/quote_type_test.go:158-176` — test constructs private fields since same package). Decide
   whether `Value()` escapes to tests only (exported) or stays unexported-method.
2. Whether handler keeps thin aliases to `representation` types or swagger annotations get updated.
3. Inline-FK-subquery failure mapping: 0-rows → `ErrNotFound` vs FK-constraint error path per
   `PersistTo`/`UpdateIn` — verify per query whether `RETURNING`/rows-affected checks are needed.
4. `PersistedUserPreference.userID` (parent's internal id as struct field): RESOLVED — legitimate
   under the persisted-owns-identity model; no action needed, but never render it.
5. `job.RunWith` signature: `func (j *PersistedProcessingJob) RunWith(ctx context.Context, p Persister, proc JobProcessor) error` — job feeds its unexported `documentID`/`documentExternalID` + type into proc internally.
6. Multiple representations per type (v1/v2 API, CSV): one `RenderX()` method per representation when
   needed — not needed now.
7. Whether `RateTable` gets `Prefetch`/`ConvertBatch` — depends on MultiProvider surface already built
   for currency work (see `handoff.md`).
8. Core-extraction granularity: per aggregate, decide whether to split an embedded core struct
   (`budget`) or keep fields flat on `Persisted*` — embedding is preferred (core-only tests, clean
   render base), but a flat struct is acceptable for tiny types.
9. Guard against promoted `Render()`: only `Persisted*` and decorator types define public
   `Render()`; cores and internals keep `render()` unexported (Go embedding would otherwise promote
   a core render that omits identity).

---

## 7. Quick file map

- Domain: `core/internal/domain/persistible.go` (1418 lines — group/category/budget/expenses/preference),
  `invitation.go`, `document.go`, `onboarding.go`, `exchange_rate.go`, plus `currency`-related types
  in `core/internal/currency/` (unchanged by this effort except RateTable impl may live there).
- Handlers: `core/internal/handler/{budget,category,expense,invitation,document,onboarding,preference}_handler.go`, `dto.go`, `expense_helpers.go`.
- Worker: `core/cmd/worker/main.go`.
- Persistence adapter: `core/internal/database/` (`PgxPersister`, `WithPersister`) — interface
  unchanged; IDs bind as plain args (named int64 types pass through pgx unchanged).
- Rules doc: `core/AGENTS.md`.
