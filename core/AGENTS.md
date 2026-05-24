# agents.md — Go Backend (Budget Management System)

## Purpose

This document defines how AI agents and developers should interact with the Go backend codebase. It enforces architectural consistency, performance constraints, and safe patterns when generating or modifying code.

The backend is a **high-performance REST API** built in Go using:

* Gin (HTTP framework)
* PostgreSQL (storage)
* SQL (preferred) or minimal ORM usage
* Event-driven + concurrent design patterns

---

## Core Principles

### 1. Explicitness over Magic

* Avoid hidden behavior, reflection-heavy patterns, or implicit side effects.
* Prefer clear data flow and explicit dependency injection.

### 2. No Global State

* DO NOT introduce global variables.
* All dependencies must be injected via constructors.

### 3. Context Propagation is Mandatory

* Every request must propagate `context.Context`.
* Never use `context.Background()` inside handlers or services.
* Respect cancellation and timeouts across DB and external calls.

### 4. Performance First

* Avoid N+1 queries.
* Prefer batch operations.
* Use connection pooling properly.
* Be mindful of allocations (consider `sync.Pool` when justified).

### 5. Predictable Concurrency

* Use worker pools for bounded concurrency.
* Avoid spawning unbounded goroutines.
* All goroutines must have lifecycle control (via context).

---

## Project Structure

```
/cmd                # entrypoints
/internal
    /api            # HTTP handlers (Gin)
    /service        # business logic
    /repository     # database access
    /model          # domain models
    /dto            # request/response objects
    /worker         # worker pools, async jobs
    /middleware     # HTTP middleware
    /config         # configuration
```

---

## Layer Responsibilities

### API Layer (`/internal/api`)

* Parse/validate requests
* Call services
* Return HTTP responses
* NO business logic

### Service Layer (`/internal/service`)

* Core business logic
* Transaction boundaries
* Coordinates repositories and workers

### Repository Layer (`/internal/repository`)

* Pure DB access
* No business logic
* Accept `context.Context`

---

## Coding Rules for Agents

### Handlers

* Always:

    * Extract `ctx := c.Request.Context()`
    * Validate input DTOs
    * Call service layer

**Example:**

```go
func (h *ExpenseHandler) Create(c *gin.Context) {
    ctx := c.Request.Context()

    var req dto.CreateExpenseRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

    res, err := h.service.CreateExpense(ctx, req)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusCreated, res)
}
```

---

### Services

* Must accept `context.Context`
* Must NOT perform direct SQL
* Must define transaction boundaries

**Example:**

```go
func (s *ExpenseService) CreateExpense(ctx context.Context, req dto.CreateExpenseRequest) (*dto.ExpenseResponse, error) {
    return s.repo.WithTx(ctx, func(txRepo repository.Repository) error {
        return txRepo.CreateExpense(ctx, req)
    })
}
```

---

### Repositories

* Only SQL / DB logic
* Always accept context
* Return domain models (not DTOs)

**Example:**

```go
func (r *ExpenseRepository) CreateExpense(ctx context.Context, req dto.CreateExpenseRequest) error {
    query := `INSERT INTO expenses (...) VALUES (...)`
    _, err := r.db.ExecContext(ctx, query, ...)
    return err
}
```

---

## Transactions

* Transactions MUST be controlled at the service layer.
* Use a `WithTx` pattern:

```go
func (r *Repo) WithTx(ctx context.Context, fn func(Repo) error) error
```

* Never nest uncontrolled transactions.

---

## Worker Pools (Required for Async Work)

Use worker pools for:

* Expense processing
* Notifications
* Heavy computations

### Rules:

* Fixed number of workers
* Bounded job queue
* Context-aware shutdown

**Example Pattern:**

```go
type Job struct {
    Ctx context.Context
    Payload any
}

type WorkerPool struct {
    jobs chan Job
}
```

---

## Memory Optimization

Use `sync.Pool` ONLY when:

* Objects are frequently allocated
* Proven hotspot (profiling required)

DO NOT prematurely optimize.

---

## Error Handling

* Use typed errors:

    * `ErrNotFound`
    * `ErrInvalidInput`
    * `ErrConflict`

* Map errors → HTTP in API layer only

---

## Logging

* Structured logging only (no fmt.Println)
* Include:

    * request_id
    * user_id
    * operation

---

## Security Rules

* Never log sensitive data (amounts will be encrypted later)
* Validate all inputs
* Use prepared statements (avoid SQL injection)

---

## Anti-Patterns (STRICTLY FORBIDDEN)

❌ Global variables
❌ context.Background() in request flow
❌ Business logic in handlers
❌ SQL inside services
❌ Unbounded goroutines
❌ Ignoring context cancellation
❌ N+1 queries

---

## Testing Guidelines

* Unit test services (mock repositories)
* Integration test repositories (real DB or test container)
* Avoid testing Gin directly unless necessary

---

## When Generating New Features

Agents MUST:

1. Create DTOs first
2. Define service interface
3. Implement repository methods
4. Wire handler last

---

## Example Feature Flow (Actual Expense)

1. `POST /expenses`
2. Handler parses request
3. Service validates + starts transaction
4. Repository inserts data
5. Worker pool optionally processes async jobs

---

## Final Rule

If unsure:
→ Choose the simplest solution that respects:

* context propagation
* no globals
* clear layering
* bounded concurrency
