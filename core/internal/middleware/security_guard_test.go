package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/budgets/core/internal/domain"
)

// boolPersister answers QueryRow by filling the first destination (a *bool)
// with result and records the queries it received.
type boolPersister struct {
	result  bool
	err     error
	queries []string
}

func (f *boolPersister) QueryRow(_ context.Context, dest []any, query string, _ ...any) error {
	f.queries = append(f.queries, query)
	if f.err != nil {
		return f.err
	}
	if b, ok := dest[0].(*bool); ok {
		*b = f.result
	}
	return nil
}

func (f *boolPersister) Exec(_ context.Context, _ string, _ ...any) (int64, error) {
	return 0, f.err
}

func (f *boolPersister) QueryRows(_ context.Context, _ func() []any, _ string, _ ...any) error {
	return f.err
}

// Regression: non-owner members must not pass the ownership check used by
// group deletion (handoff Phase 0, item 6).
func TestAuthorizeGroupOwnership_MemberIsForbidden(t *testing.T) {
	p := &boolPersister{result: false}
	guard := NewSecurityGuard(&domain.PersistedUser{})

	err := guard.AuthorizeGroupOwnership(context.Background(), p, uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestAuthorizeGroupOwnership_OwnerAllowed(t *testing.T) {
	p := &boolPersister{result: true}
	guard := NewSecurityGuard(&domain.PersistedUser{})

	err := guard.AuthorizeGroupOwnership(context.Background(), p, uuid.New())

	assert.NoError(t, err)
}

// The ownership check must filter on role='owner'; a plain membership check
// would let members delete groups.
func TestAuthorizeGroupOwnership_QueryRequiresOwnerRole(t *testing.T) {
	p := &boolPersister{result: true}
	guard := NewSecurityGuard(&domain.PersistedUser{})

	require.NoError(t, guard.AuthorizeGroupOwnership(context.Background(), p, uuid.New()))
	require.Len(t, p.queries, 1)
	assert.Contains(t, p.queries[0], "up.role = 'owner'")
}

func TestAuthorizeGroupOwnership_PersisterErrorPropagates(t *testing.T) {
	boom := errors.New("connection reset")
	p := &boolPersister{err: boom}
	guard := NewSecurityGuard(&domain.PersistedUser{})

	err := guard.AuthorizeGroupOwnership(context.Background(), p, uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
	assert.NotErrorIs(t, err, domain.ErrForbidden)
}

func TestAuthorizeGroupAccess_NonMemberIsForbidden(t *testing.T) {
	p := &boolPersister{result: false}
	guard := NewSecurityGuard(&domain.PersistedUser{})

	err := guard.AuthorizeGroupAccess(context.Background(), p, uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}
