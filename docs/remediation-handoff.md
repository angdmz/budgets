# Core Architecture Remediation — Handoff

Plan to bring `core/` into compliance with `core/AGENTS.md`, plus regression
tests for the bugs found during evaluation.

Each phase is independently mergeable. Phases are ordered by risk × value:
bugs first, then boundary violations, then design-rule violations.

---

## Phase 0 — Regression tests for found bugs

Add failing tests first; they pass after the phases below fix the code.
Tests live in `tests/` (Python integration suite) unless noted.

1. **Invitation list `accepted_at` always null**
   - Bug: `PersistedInvitationsForGroup` scans `acceptedByUserID`/`acceptedAt`
     into locals never assigned to the struct
     (`core/internal/domain/invitation.go:216-237`).
   - Test: create group → invite → accept with second user → owner lists
     invitations → assert `accepted_at` is non-null.
2. **DB errors masked as 404**
   - Bug: every `FromPersistence` wraps all errors as `ErrNotFound`;
     `PgxPersister` never translates `pgx.ErrNoRows`
     (`core/internal/database/persister.go:18-20`).
   - Test (Go unit test in `core/`, new): fake `Persister` returning a
     non-`ErrNoRows` error → assert `errors.Is(err, ErrNotFound)` is false and
     the original error propagates. Also assert `ErrNoRows` → `ErrNotFound`.
3. **Expenses list N+1 provider calls**
   - Bug: `ConvertHistorical` bypasses cache; list endpoints call it per
     expense (`currency/provider.go:86-98`,
     `handler/expense_handler.go:240-258`, `599-615`,
     `domain/expense_render.go:187-201`).
   - Test (Go): stub `ExchangeRateProvider` counting `GetHistoricalRate`
     calls → list 3 expenses in a foreign currency → assert ≤1 call per
     distinct (from,to,quote,date) tuple.
4. **`PersistibleGroup.AddCategory` data silently dropped**
   - Bug: `categories`/`budgets` slices are never persisted in `PersistTo`
     (`domain/persistible.go:56-121`); `AddCategory` skips name validation.
   - Decision needed: either wire persistence or delete the dead collection
     API. Test whichever is chosen.
5. **`CreateGroup` returns 500 for validation errors**
   - Bug: all errors → `SafeErrorResponse(500)` (`group_handler.go:71-74`).
   - Test: POST `/groups` with empty name → expect 400.
6. **Non-owner member can delete a group**
   - Bug: `DeleteGroup` uses `AuthorizeGroupAccess`, not ownership
     (`group_handler.go:279-281`).
   - Test: member (non-owner) DELETE `/groups/{id}` → expect 403.

## Phase 1 — Correctness fixes (no design change)

- Fix invitation `acceptedAt`/`acceptedByUserID` scan (assign locals to struct).
- `PgxPersister.QueryRow`: translate `pgx.ErrNoRows` → `domain.ErrNotFound`
  (wrap, not bare). Remove blanket `ErrNotFound` wrapping in `FromPersistence`
  functions so real DB errors propagate.
- `DeleteGroup` → `AuthorizeGroupOwnership`; order checks authorize-first in
  `RevokeInvitation` (authorize before fetching the invitation to avoid the
  404/403 existence oracle).
- Map `ErrValidation` in `CreateGroup`/`CreateCategory`-style error blocks
  (or reuse `handleServiceError`).
- Remove `[DEBUG]` log lines from expense handlers.

## Phase 2 — O(1) I/O for expense lists & summary

Goal (AGENTS §3.7 of handoff doc, rule line 18): fetch needed rates once per
request, never per-expense.

- Extend `ExchangeRateCache` key with `date` (or add a historical-rate cache)
  so `ConvertHistorical` can cache. Alternatively add
  `Marketplace.RatesFor(pairs, asOf)` that batch-fetches distinct tuples once.
- `NewBudgetSummary` and list rendering: collect distinct
  (currency,quote,date) tuples, fetch each once, then build renderings.
- Keep decryption+conversion per expense (CPU-only) but ensure zero network
  calls inside the loop after batch fetch.
- Gate remaining latency with bounded concurrency if batch API is
  unavailable (rule line 11) — likely unnecessary once caching lands.

## Phase 3 — Persistence boundary

- Move handler SQL into domain: `invitation_handler.go:48-64` and `:110-119`
  → `PersistedGroupFromPersistence` / a domain `GroupExists`-style check is
  NOT wanted (getter); instead have `NewPersistibleInvitation` take the
  group **external** ID and resolve the internal id via inline subquery in
  the INSERT (rule line 33).
- Same inline-subquery change for `PersistibleCategory`, `PersistibleBudget`,
  `PersistibleExpectedExpense`, `PersistibleActualExpense`
  (`INSERT ... (SELECT id FROM ... WHERE external_id = $N AND revoked_at IS NULL)`),
  and their `UpdateIn` methods; drop the extra `SELECT id` round trips in the
  `*ForGroup`/`*ForBudget` loaders by joining on external id directly.
- Introduce named internal id types (`groupID`, `budgetID`, …) or keep them
  encapsulated but never param-typed as bare `int64` crossing the package
  boundary (rule line 31).
- Move `SecurityGuard` out of `domain` (to `middleware` or a `guard` package);
  domain objects must not contain authorization (rule line 37).

## Phase 4 — Rendering seal & tell-don't-ask

- Unexport `Render` → `render` and `Rendered.Value` → unexported or remove;
  handlers only ever pass `Rendered[T]` to `c.JSON` (verify no external
  callers; grep shows none today).
- `Money`: unexport fields (add `NewMoney(amount, currency)` already exists);
  replace field reads (`money.Currency` in `PresentationPrefs.Convert`,
  `renderMoney`, `CurrencyMarketplace`) with behavior methods
  (`m.CurrencyIs(c)`, `m.AmountIn(...)`) — or minimally keep fields but drop
  JSON tags and route reads through methods to satisfy rule 6.
- Onboarding handlers → single domain verbs: `ob.CompleteStep(step, data, p)`,
  `ob.SkipStep(step, p)`, `ob.GoBackFrom(step, p)` — absorbing `Next()/Prev()`
  branching (rule 10).
- Expense list handlers: replace `if pErr == nil` preference fetch with a
  NullObject (`PresentationPrefs` obtained via
  `PersistedUserPreferenceOrDefault(ctx, userID, p)`); likewise null
  `MoneyConverter` instead of `m == nil` checks; remove
  `expectedDate.IsZero()` branching by making the render pipeline handle
  absence internally.
- `ConversionCutoffAt` is a getter — fold it into `BudgetSummary`
  construction (already internal) and delete the public method if unused.

## Phase 5 — NullObject & polymorphism

- Split `PersistibleUserOnboardingStep` into `CompletedStep`/`SkippedStep`
  persistibles; delete `PersistSkippedTo` (rule 15).
- Invitation state: pending/accepted/revoked as distinct types or internal
  state machine so `Accept`/`Revoke`/`EnsureUsable` don't string-branch.
  Minimal acceptable step: typed `InvitationStatus` with behavior methods.
- `validateStepData` switch → per-step validator types, or table of
  step→validator.
- `AddUser(isPrimary bool)` → `AddPrimaryUser`/`AddMemberUser` (drop boolean
  flag).
- Find-or-create constructors: `PersistedUserOnboardingFor(ctx, userID, p)`
  and `PersistedUserPreferenceFor(ctx, userID, p)` that internally create
  defaults on `ErrNotFound` — removes the 5×/2× duplicated blocks in handlers.

## Phase 6 — Dependency injection (rule 4)

- Inject clock: `NewPersistibleInvitation`, `Accept`, `NewBudgetSummary`,
  `InMemoryCache` take `now time.Time` or a `Clock` interface.
- Inject randomness: token generation takes an `io.Reader` (default
  `crypto/rand.Reader`).
- `Money` JSON tags: domain types must not marshal to wire (rule 9);
  `representation.Money` is already the wire type — drop tags.

## Phase 7 — Legacy cleanup (rule 40)

- `domain.User`/`BaseModel` still used by `repository/`, `middleware/`,
  `auth_handler.go`. Migrate user resolution to `Persisted*` objects; keep a
  thin `middleware` DTO only if needed for gin context (evaluate whether it
  violates the boundary).
- Decide `PersistedExchangeRate` fate: it has no render path; either give it
  a render/usage or remove it (it's currently dead weight).
- `OnboardingStepStatus`/status constants: consolidate stringly-typed status
  enums as part of Phase 5.

## Open questions for the user

1. `PersistibleGroup.AddCategory`/`AddBudget`: keep & persist, or delete?
2. Currency DB `PersistedExchangeRateFromPersistence`: intended for a future
   provider-backed-by-DB, or deletable?
3. Preferred fix for historical rates: cache-by-date vs batch fetch API?
4. Should non-expense lists (invitations) also gain `accepted_by` rendering
   or is exposing internal `acceptedByUserID` unnecessary (privacy)?

## Verification

- `cd core && go build ./... && go vet ./...` after each phase.
- Run `tests/` suite (Docker) for Phases 0–3 — integration tests cover the
  bug fixes.
- New Go unit tests under `core/internal/...` for persister error
  translation and marketplace call-counting.
