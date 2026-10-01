package domain

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression: PersistedInvitationsForGroup must scan accepted_by_user_id and
// accepted_at into the persisted struct so they reach the wire shape
// (handoff Phase 0, item 1).
func TestPersistedInvitationsForGroup_ScansAcceptedFields(t *testing.T) {
	acceptedBy := int64(42)
	acceptedAt := time.Date(2026, 9, 20, 15, 4, 5, 0, time.UTC)
	expiresAt := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := acceptedAt
	invExternalID := uuid.New()

	p := &fakePersister{rows: [][]any{{
		int64(1), invExternalID, int64(7), "Family", int64(9), "Owner",
		&acceptedBy, "tok123", InvitationStatusAccepted, ParticipantRoleMember,
		expiresAt, &acceptedAt, createdAt, updatedAt,
	}}}

	invs, err := PersistedInvitationsForGroup(context.Background(), uuid.New(), p)

	require.NoError(t, err)
	require.Len(t, invs, 1)

	rendered := invs[0].Render().value()
	require.NotNil(t, rendered.AcceptedAt)
	assert.Equal(t, acceptedAt, *rendered.AcceptedAt)
	assert.Equal(t, "accepted", rendered.Status)
	assert.Equal(t, invExternalID, rendered.ID)
}

// sequencePersister answers QueryRow calls in order with the configured
// fills; Accept issues EXISTS -> participant INSERT -> invitation UPDATE.
type sequencePersister struct {
	fakePersister
	fills   [][]any
	callIdx int
}

func (s *sequencePersister) QueryRow(ctx context.Context, dest []any, q string, args ...any) error {
	if s.rowErr != nil {
		return s.rowErr
	}
	if s.callIdx < len(s.fills) {
		fillDest(dest, s.fills[s.callIdx])
	}
	s.callIdx++
	return nil
}

func TestPersistedInvitation_AcceptSetsAcceptedAt(t *testing.T) {
	now := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	expiresAt := now.Add(24 * time.Hour)
	userID := int64(42)

	p := &sequencePersister{fills: [][]any{
		{false},                                  // alreadyMember check
		{int64(10), uuid.New(), now, now},        // participant insert
		{now},                                    // invitation update
	}}

	inv := &PersistedInvitation{
		id:        1,
		groupID:   7,
		token:     "tok",
		status:    InvitationStatusPending,
		role:      ParticipantRoleMember,
		expiresAt: expiresAt,
	}

	err := inv.Accept(context.Background(), &PersistedUser{id: userID, displayName: "New Member"}, now, p)
	require.NoError(t, err)
	assert.Equal(t, InvitationStatusAccepted, inv.status)
	require.NotNil(t, inv.acceptedAt)
	assert.Equal(t, now, *inv.acceptedAt)
	require.NotNil(t, inv.acceptedByUserID)
	assert.Equal(t, userID, *inv.acceptedByUserID)
}

func TestPersistedInvitation_AcceptAlreadyAcceptedConflicts(t *testing.T) {
	now := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	inv := &PersistedInvitation{
		status:    InvitationStatusAccepted,
		expiresAt: now.Add(24 * time.Hour),
	}

	err := inv.Accept(context.Background(), &PersistedUser{id: 42}, now, &fakePersister{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConflict)
}
