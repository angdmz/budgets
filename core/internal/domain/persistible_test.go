package domain

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePersister is a Persister stub that fills scan destinations from
// configured row values and returns configured errors.
type fakePersister struct {
	rowFill  []any
	rowErr   error
	rows     [][]any
	rowsErr  error
	execRows int64
	execErr  error
}

func (f *fakePersister) QueryRow(_ context.Context, dest []any, _ string, _ ...any) error {
	if f.rowErr != nil {
		return f.rowErr
	}
	fillDest(dest, f.rowFill)
	return nil
}

func (f *fakePersister) Exec(_ context.Context, _ string, _ ...any) (int64, error) {
	return f.execRows, f.execErr
}

func (f *fakePersister) QueryRows(_ context.Context, dest func() []any, _ string, _ ...any) error {
	for _, row := range f.rows {
		fillDest(dest(), row)
	}
	return f.rowsErr
}

func fillDest(dest []any, values []any) {
	for i := range dest {
		if i < len(values) && values[i] != nil {
			reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(values[i]))
		}
	}
}

// Regression: real DB errors must propagate instead of being masked as
// ErrNotFound (handoff Phase 0, item 2).
func TestPersistedGroupFromPersistence_OtherErrorsPropagate(t *testing.T) {
	boom := errors.New("connection reset")
	p := &fakePersister{rowErr: boom}

	_, err := PersistedGroupFromPersistence(context.Background(), uuid.New(), p)

	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
	assert.NotErrorIs(t, err, ErrNotFound)
}

func TestPersistedGroupFromPersistence_NotFoundStillMaps(t *testing.T) {
	p := &fakePersister{rowErr: ErrNotFound}

	_, err := PersistedGroupFromPersistence(context.Background(), uuid.New(), p)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestNewPersistibleGroup_EmptyNameIsValidationError(t *testing.T) {
	_, err := NewPersistibleGroup("", "desc")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrValidation)
}
