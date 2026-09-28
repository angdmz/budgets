package domain

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/budgets/core/internal/representation"
)

const (
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusRevoked  = "revoked"
	InvitationStatusExpired  = "expired"
)

type PersistibleInvitation struct {
	groupID       int64
	inviterUserID int64
	token         string
	role          string
	expiresAt     time.Time
}

func NewPersistibleInvitation(groupID int64, inviterUserID int64, role string) (*PersistibleInvitation, error) {
	if role == "" {
		role = "member"
	}

	token, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	return &PersistibleInvitation{
		groupID:       groupID,
		inviterUserID: inviterUserID,
		token:         token,
		role:          role,
		expiresAt:     expiresAt,
	}, nil
}

func (i *PersistibleInvitation) PersistTo(ctx context.Context, p Persister) (*PersistedInvitation, error) {
	var id int64
	var externalID uuid.UUID
	var groupName, inviterName string
	var createdAt, updatedAt time.Time

	var groupExternalID uuid.UUID

	err := p.QueryRow(
		ctx,
		[]any{&id, &externalID, &groupExternalID, &groupName, &inviterName, &createdAt, &updatedAt},
		`INSERT INTO group_invitations (budgeting_group_id, inviter_user_id, token, role, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, external_id,
			(SELECT external_id FROM budgeting_groups WHERE id = $1),
			(SELECT name FROM budgeting_groups WHERE id = $1),
			(SELECT display_name FROM users WHERE id = $2),
			created_at, updated_at`,
		i.groupID, i.inviterUserID, i.token, i.role, InvitationStatusPending, i.expiresAt,
	)
	if err != nil {
		return nil, err
	}

	return &PersistedInvitation{
		id:              id,
		externalID:      externalID,
		groupID:         i.groupID,
		groupExternalID: groupExternalID,
		groupName:       groupName,
		inviterUserID:   i.inviterUserID,
		inviterName:     inviterName,
		token:           i.token,
		status:          InvitationStatusPending,
		role:            i.role,
		expiresAt:       i.expiresAt,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}, nil
}

type PersistedInvitation struct {
	id               int64
	externalID       uuid.UUID
	groupID          int64
	groupExternalID  uuid.UUID
	groupName        string
	inviterUserID    int64
	inviterName      string
	acceptedByUserID *int64
	token            string
	status           string
	role             string
	expiresAt        time.Time
	acceptedAt       *time.Time
	createdAt        time.Time
	updatedAt        time.Time
}

func PersistedInvitationByToken(ctx context.Context, token string, p Persister) (*PersistedInvitation, error) {
	var inv PersistedInvitation
	var acceptedByUserID *int64
	var acceptedAt *time.Time

	err := p.QueryRow(
		ctx,
		[]any{
			&inv.id,
			&inv.externalID,
			&inv.groupID,
			&inv.groupExternalID,
			&inv.groupName,
			&inv.inviterUserID,
			&inv.inviterName,
			&acceptedByUserID,
			&inv.token,
			&inv.status,
			&inv.role,
			&inv.expiresAt,
			&acceptedAt,
			&inv.createdAt,
			&inv.updatedAt,
		},
		`SELECT
			gi.id, gi.external_id, gi.budgeting_group_id, bg.external_id,
			bg.name, gi.inviter_user_id, u.display_name,
			gi.accepted_by_user_id, gi.token, gi.status, gi.role,
			gi.expires_at, gi.accepted_at, gi.created_at, gi.updated_at
		FROM group_invitations gi
		JOIN budgeting_groups bg ON gi.budgeting_group_id = bg.id
		JOIN users u ON gi.inviter_user_id = u.id
		WHERE gi.token = $1 AND gi.revoked_at IS NULL`,
		token,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: invitation not found", ErrNotFound)
	}

	inv.acceptedByUserID = acceptedByUserID
	inv.acceptedAt = acceptedAt

	return &inv, nil
}

func PersistedInvitationByExternalID(ctx context.Context, externalID uuid.UUID, p Persister) (*PersistedInvitation, error) {
	var inv PersistedInvitation
	var acceptedByUserID *int64
	var acceptedAt *time.Time

	err := p.QueryRow(
		ctx,
		[]any{
			&inv.id,
			&inv.externalID,
			&inv.groupID,
			&inv.groupExternalID,
			&inv.groupName,
			&inv.inviterUserID,
			&inv.inviterName,
			&acceptedByUserID,
			&inv.token,
			&inv.status,
			&inv.role,
			&inv.expiresAt,
			&acceptedAt,
			&inv.createdAt,
			&inv.updatedAt,
		},
		`SELECT
			gi.id, gi.external_id, gi.budgeting_group_id, bg.external_id,
			bg.name, gi.inviter_user_id, u.display_name,
			gi.accepted_by_user_id, gi.token, gi.status, gi.role,
			gi.expires_at, gi.accepted_at, gi.created_at, gi.updated_at
		FROM group_invitations gi
		JOIN budgeting_groups bg ON gi.budgeting_group_id = bg.id
		JOIN users u ON gi.inviter_user_id = u.id
		WHERE gi.external_id = $1 AND gi.revoked_at IS NULL`,
		externalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: invitation not found", ErrNotFound)
	}

	inv.acceptedByUserID = acceptedByUserID
	inv.acceptedAt = acceptedAt

	return &inv, nil
}

func PersistedInvitationsForGroup(ctx context.Context, groupExternalID uuid.UUID, p Persister) ([]PersistedInvitation, error) {
	var groupID int64
	err := p.QueryRow(
		ctx,
		[]any{&groupID},
		`SELECT id FROM budgeting_groups WHERE external_id = $1 AND revoked_at IS NULL`,
		groupExternalID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: group not found", ErrNotFound)
	}

	invitations := make([]PersistedInvitation, 0)
	err = p.QueryRows(
		ctx,
		func() []any {
			var inv PersistedInvitation
			var acceptedByUserID *int64
			var acceptedAt *time.Time
			inv.groupExternalID = groupExternalID
			invitations = append(invitations, inv)
			idx := len(invitations) - 1
			return []any{
				&invitations[idx].id,
				&invitations[idx].externalID,
				&invitations[idx].groupID,
				&invitations[idx].groupName,
				&invitations[idx].inviterUserID,
				&invitations[idx].inviterName,
				&acceptedByUserID,
				&invitations[idx].token,
				&invitations[idx].status,
				&invitations[idx].role,
				&invitations[idx].expiresAt,
				&acceptedAt,
				&invitations[idx].createdAt,
				&invitations[idx].updatedAt,
			}
		},
		`SELECT
			gi.id, gi.external_id, gi.budgeting_group_id,
			bg.name, gi.inviter_user_id, u.display_name,
			gi.accepted_by_user_id, gi.token, gi.status, gi.role,
			gi.expires_at, gi.accepted_at, gi.created_at, gi.updated_at
		FROM group_invitations gi
		JOIN budgeting_groups bg ON gi.budgeting_group_id = bg.id
		JOIN users u ON gi.inviter_user_id = u.id
		WHERE gi.budgeting_group_id = $1 AND gi.revoked_at IS NULL
		ORDER BY gi.created_at DESC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}

	return invitations, nil
}

// Render returns the final wire representation of this invitation.
func (i *PersistedInvitation) Render() Rendered[representation.Invitation] {
	return Render(representation.Invitation{
		ID:          i.externalID,
		Token:       i.token,
		GroupID:     i.groupExternalID,
		GroupName:   i.groupName,
		InviterName: i.inviterName,
		Status:      i.status,
		Role:        i.role,
		ExpiresAt:   i.expiresAt,
		AcceptedAt:  i.acceptedAt,
		CreatedAt:   i.createdAt,
	})
}

// RenderDetail returns the public wire representation of this invitation.
func (i *PersistedInvitation) RenderDetail() Rendered[representation.InvitationDetail] {
	return Render(representation.InvitationDetail{
		GroupName:   i.groupName,
		InviterName: i.inviterName,
		Status:      i.status,
		Role:        i.role,
		ExpiresAt:   i.expiresAt,
	})
}

// EnsureUsable fails with ErrGone when this invitation is revoked or has
// already expired at the given instant.
func (i *PersistedInvitation) EnsureUsable(now time.Time) error {
	if i.status == InvitationStatusRevoked {
		return fmt.Errorf("%w: invitation has been revoked", ErrGone)
	}
	if now.After(i.expiresAt) {
		return fmt.Errorf("%w: invitation has expired", ErrGone)
	}
	return nil
}

func (i *PersistedInvitation) Accept(ctx context.Context, userID int64, displayName string, p Persister) error {
	if i.status == InvitationStatusRevoked {
		return fmt.Errorf("%w: invitation has been revoked", ErrGone)
	}

	if i.status != InvitationStatusPending {
		return fmt.Errorf("%w: invitation already %s", ErrConflict, i.status)
	}

	if time.Now().After(i.expiresAt) {
		return fmt.Errorf("%w: invitation has expired", ErrGone)
	}

	var alreadyMember bool
	err := p.QueryRow(
		ctx,
		[]any{&alreadyMember},
		`SELECT EXISTS(
			SELECT 1 FROM user_participants up
			JOIN participants pt ON up.participant_id = pt.id
			WHERE up.user_id = $1 AND pt.budgeting_group_id = $2
			AND up.revoked_at IS NULL AND pt.revoked_at IS NULL
		)`,
		userID, i.groupID,
	)
	if err != nil {
		return err
	}

	if alreadyMember {
		return fmt.Errorf("%w: user is already a member of this group", ErrConflict)
	}

	var participantID int64
	var participantExternalID uuid.UUID
	var pCreatedAt, pUpdatedAt time.Time

	err = p.QueryRow(
		ctx,
		[]any{&participantID, &participantExternalID, &pCreatedAt, &pUpdatedAt},
		`INSERT INTO participants (name, description, budgeting_group_id)
		VALUES ($1, $2, $3)
		RETURNING id, external_id, created_at, updated_at`,
		displayName, "", i.groupID,
	)
	if err != nil {
		return err
	}

	isPrimaryInt := 0
	_, err = p.Exec(
		ctx,
		`INSERT INTO user_participants (user_id, participant_id, role, is_primary)
		VALUES ($1, $2, $3, $4)`,
		userID, participantID, i.role, isPrimaryInt,
	)
	if err != nil {
		return err
	}

	now := time.Now()
	var updatedAt time.Time
	err = p.QueryRow(
		ctx,
		[]any{&updatedAt},
		`UPDATE group_invitations
		SET status = $1, accepted_by_user_id = $2, accepted_at = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING updated_at`,
		InvitationStatusAccepted, userID, now, i.id,
	)
	if err != nil {
		return err
	}

	i.status = InvitationStatusAccepted
	i.acceptedByUserID = &userID
	i.acceptedAt = &now
	i.updatedAt = updatedAt

	return nil
}

func (i *PersistedInvitation) Revoke(ctx context.Context, p Persister) error {
	if i.status != InvitationStatusPending {
		return fmt.Errorf("%w: can only revoke pending invitations", ErrConflict)
	}

	var updatedAt time.Time
	err := p.QueryRow(
		ctx,
		[]any{&updatedAt},
		`UPDATE group_invitations
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING updated_at`,
		InvitationStatusRevoked, i.id,
	)
	if err != nil {
		return err
	}

	i.status = InvitationStatusRevoked
	i.updatedAt = updatedAt

	return nil
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
