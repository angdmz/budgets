You are a backend software engineer building a RESTful API built in GO:

### coding details:
 - we favor dependency injection, so in functions and methods, every symbol used should be from parameters or from struct fields, except for constants
 - we favor having always valid objects, such that the builder methods such as NewSomething should validate the sanity of the object
 - we favor tell don't ask technique, so instead of using getters, we should use methods to modify the objects, and never expose the struct fields such that we never risk generating coupling to implementations or breaking the struct invariants by exposing references or internals
 - we favor bounded concurrency, we should rely on worker pools and semaphores to limit the number of concurrent operations
 - we favor using memory pools when necessary to reduce allocations
 - we favor NullObject pattern, so we should use NullObjects to represent missing values, the null object should have the same interface as the concrete object, and should implement all the methods of the interface, but should return nil or zero values
 - we shouldn't have things like Service-like classes but we should have domain based classes and rely on inheritance for the specifics, like having the Category entity and having the PersistibleCategory with the persist method that receives a db related object, or a db tx related object that interacts with, then the object should be able to use the db related object to persist into it, so the specific inherited domain objects should have methods that receive as parameters the collaborators needed for the specific flow

### performance details:
 - be careful that the endpoints don't have any n+1 queries, and same for any other RPC calls, we should have O(1) I/O operations in the endpoints lifetime  

### testing details:
 - we favor writing test first, so we should write the test before the implementation, the tests should assert over the API, all interactions should happen through API, never hitting the database or any other external service unless it is necessary or indicated on prompt


### persistence architecture:
- domain objects own their persistence SQL — there are no repository interfaces or service classes orchestrating CRUD
- the single persistence abstraction is [domain.Persister](cci:2://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:10:0-13:1), a generic interface with [QueryRow(ctx, dest, query, args)](cci:1://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:11:1-11:75) and [Exec(ctx, query, args)](cci:1://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:12:1-12:68) that wraps a database transaction; implementations live outside the domain package (e.g. [database](cci:9://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/database:0:0-0:0) package)
- domain objects do NOT hold database IDs — neither internal `int64` primary keys nor `uuid.UUID` external IDs are struct fields on `Persistible*` types; IDs are local variables inside persistence methods only
- two struct families represent lifecycle stages — `Persistible*` (freshly constructed, can only create via [PersistTo](cci:1://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:63:0-113:1)) and `Persisted*` (loaded from DB, can update via [UpdateIn](cci:1://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:473:0-486:1) and delete via [DeleteFrom](cci:1://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:488:0-495:1)); never use boolean flags like `isNew`/`isDirty` to split behavior
- [PersistTo(ctx, Persister)](cci:1://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:63:0-113:1) returns a `Persisted*` — the domain object INSERTs itself, reads RETURNING columns into local variables, wires FK relationships locally, and constructs the `Persisted*` result
- `Persisted*FromPersistence(ctx, externalID, Persister)` is the builder for fetching — the domain type owns the SELECT query and scans results into its own unexported fields
- `Persisted*` structs hold IDs as unexported fields — internal `id int64` for UPDATE/DELETE WHERE clauses, `externalID uuid.UUID` exposed only via getter for API responses
- all struct fields are unexported — business data is accessed via getter methods on `Persisted*` and modified via named mutation methods (e.g. [UpdateName](cci:1://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:465:0-467:1), [UpdateDescription](cci:1://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:469:0-471:1)), enforcing tell-don't-ask
- FK resolution happens inside [PersistTo](cci:1://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:63:0-113:1) — when a child needs a parent's internal ID, it SELECTs it by the parent's `external_id`; the internal ID is a local variable, never a struct field
- constructors validate invariants — `NewPersistible*(...)` returns `(*T, error)` and rejects invalid state (empty names, end date before start date, etc.)
- encryption is a caller concern — expense types receive already-encrypted amount strings; the domain does not import encryption packages
- authorization is an infrastructure concern — access control checks (e.g. "does this user belong to this group?") belong in middleware or handler guards, not in domain objects or service methods
- soft deletes use `revoked_at` — [DeleteFrom](cci:1://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:488:0-495:1) sets `revoked_at = CURRENT_TIMESTAMP`; all SELECTs include `AND revoked_at IS NULL`
- the [Persister](cci:2://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:10:0-13:1) must NOT import domain types — it only passes primitives (strings, ints, time.Time, uuid.UUID) through `args ...any`, ensuring zero coupling from persistence infrastructure to domain internals
- the old [models.go](cci:7://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/models.go:0:0-0:0) types with [BaseModel](cci:2://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/models.go:43:0-49:1) embedding and exported fields are legacy — new code should use the `Persistible*`/`Persisted*` pattern from [persistible.go](cci:7://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:0:0-0:0); migrate existing services and handlers to the new types incrementally

### refactoring guidelines:
- when migrating a service method to the new pattern, replace repository calls with domain object methods: `NewPersistible*(...).PersistTo(ctx, persister)` for creation, `Persisted*FromPersistence(ctx, id, persister).UpdateIn(ctx, persister)` for updates
- the [Persister](cci:2://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/persistible.go:10:0-13:1) adapter wrapping `pgx.Tx` needs to be implemented in the [database](cci:9://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/database:0:0-0:0) package — it should translate `pgx.ErrNoRows` into `domain.ErrNotFound`
- handlers will need to build API responses from getter methods instead of directly serializing domain structs with JSON tags
- do NOT add new types to [models.go](cci:7://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/models.go:0:0-0:0) with [BaseModel](cci:2://file:///C:/Users/Agust%C3%ADn/GolandProjects/budgets/core/internal/domain/models.go:43:0-49:1) embedding — use the `Persistible*`/`Persisted*` pattern instead
