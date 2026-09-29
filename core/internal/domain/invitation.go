package domain

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/budgets/core/internal/representation"
)

// InvitationStatus is the lifecycle state of an invitation. State transitions
// are enforced by its behavior methods, not by string comparisons at call
// sites.
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusRevoked  InvitationStatus = "revoked"
	InvitationStatusExpired  InvitationStatus = "expired"
)

// ensureAcceptable fails unless the invitation can be accepted at `now`.
func (s InvitationStatus) ensureAcceptable(now, expiresAt time.Time) error {
	if s == InvitationStatusRevoked {
		return fmt.Errorf("%w: invitation has been revoked", ErrGone)
	}
	if s != InvitationStatusPending {
		return fmt.Errorf("%w: invitation already %s", ErrConflict, s)
	}
	if now.After(expiresAt) {
		return fmt.Errorf("%w: invitation has expired", ErrGone)
	}
	return nil
}

// ensureRevokable fails unless the invitation is still pending.
func (s InvitationStatus) ensureRevokable() error {
	if s != InvitationStatusPending {
		return fmt.Errorf("%w: can only revoke pending invitations", ErrConflict)
	}
	return nil
}

// ensureUsable fails when the invitation is revoked or expired at `now`.
func (s InvitationStatus) ensureUsable(now, expiresAt time.Time) error {
	if s == InvitationStatusRevoked {
		return fmt.Errorf("%w: invitation has been revoked", ErrGone)
	}
	if now.After(expiresAt) {
		return fmt.Errorf("%w: invitation has expired", ErrGone)
	}
	return nil
}

type PersistibleInvitation struct {
	groupExternalID uuid.UUID
	inviter         *PersistedUser
	token           string
	role            ParticipantRole
	expiresAt       time.Time
}

// NewPersistibleInvitation builds a new invitation for the group identified by
// its external ID. `entropy` is the source for the token (crypto/rand.Reader in
// production); `now` fixes the expiry instant.
func NewPersistibleInvitation(groupExternalID uuid.UUID, inviter *PersistedUser, role ParticipantRole, entropy io.Reader, now time.Time) (*PersistibleInvitation, error) {
	if inviter == nil {
		return nil, fmt.Errorf("%w: inviter is required", ErrValidation)
	}
	if role == "" {
		role = ParticipantRoleMember
	}

	token, err := generateSecureToken(entropy)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &PersistibleInvitation{
		groupExternalID: groupExternalID,
		inviter:         inviter,
		token:           token,
		role:            role,
		expiresAt:       now.Add(7 * 24 * time.Hour),
	}, nil
}

func (i *PersistibleInvitation) PersistTo(ctx context.Context, p Persister) (*PersistedInvitation, error) {
	var id, groupID int64
	var externalID uuid.UUID
	var groupName, inviterName string
	var createdAt, updatedAt time.Time

	err := p.QueryRow(
		ctx,
		[]any{&id, &externalID, &groupID, &groupName, &inviterName, &createdAt, &updatedAt},
		`WITH ins AS (
			INSERT INTO group_invitations (budgeting_group_id, inviter_user_id, token, role, status, expires_at)
			SELECT bg.id, $2, $3, $4, $5, $6
			FROM budgeting_groups bg
			WHERE bg.external_id = $1 AND bg.revoked_at IS NULL
			RETURNING id, external_id, budgeting_group_id, created_at, updated_at
		)
		SELECT ins.id, ins.external_id, ins.budgeting_group_id, bg.name,
			(SELECT display_name FROM users WHERE id = $2),
			ins.created_at, ins.updated_at
		FROM ins
		JOIN budgeting_groups bg ON bg.id = ins.budgeting_group_id`,
		i.groupExternalID, i.inviter, i.token, string(i.role), InvitationStatusPending, i.expiresAt,
	)
	if err != nil {
		return nil, wrapNotFound(err, "group not found")
	}

	return &PersistedInvitation{
		id:              id,
		externalID:      externalID,
		groupID:         groupID,
		groupExternalID: i.groupExternalID,
		groupName:       groupName,
		inviterUserID:   i.inviter.id,
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
	status           InvitationStatus
	role             ParticipantRole
	expiresAt        time.Time
	acceptedAt       *time.Time
	createdAt        time.Time
	updatedAt        time.Time
}

func PersistedInvitationByToken(ctx context.Context, token string, p Persister) (*PersistedInvitation, error) {
	var inv PersistedInvitation

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
			&inv.acceptedByUserID,
			&inv.token,
			&inv.status,
			&inv.role,
			&inv.expiresAt,
			&inv.acceptedAt,
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
		return nil, wrapNotFound(err, "invitation not found")
	}

	return &inv, nil
}

func PersistedInvitationByExternalID(ctx context.Context, externalID uuid.UUID, p Persister) (*PersistedInvitation, error) {
	var inv PersistedInvitation

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
			&inv.acceptedByUserID,
			&inv.token,
			&inv.status,
			&inv.role,
			&inv.expiresAt,
			&inv.acceptedAt,
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
		return nil, wrapNotFound(err, "invitation not found")
	}

	return &inv, nil
}

func PersistedInvitationsForGroup(ctx context.Context, groupExternalID uuid.UUID, p Persister) ([]PersistedInvitation, error) {
	invitations := make([]PersistedInvitation, 0)
	err := p.QueryRows(
		ctx,
		func() []any {
			var inv PersistedInvitation
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
				&invitations[idx].acceptedByUserID,
				&invitations[idx].token,
				&invitations[idx].status,
				&invitations[idx].role,
				&invitations[idx].expiresAt,
				&invitations[idx].acceptedAt,
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
		WHERE bg.external_id = $1 AND bg.revoked_at IS NULL AND gi.revoked_at IS NULL
		ORDER BY gi.created_at DESC`,
		groupExternalID,
	)
	if err != nil {
		return nil, err
	}

	return invitations, nil
}

// Render returns the final wire representation of this invitation.
func (i *PersistedInvitation) Render() Rendered[representation.Invitation] {
	return render(representation.Invitation{
		ID:          i.externalID,
		Token:       i.token,
		GroupID:     i.groupExternalID,
		GroupName:   i.groupName,
		InviterName: i.inviterName,
		Status:      string(i.status),
		Role:        string(i.role),
		ExpiresAt:   i.expiresAt,
		AcceptedAt:  i.acceptedAt,
		CreatedAt:   i.createdAt,
	})
}

// RenderDetail returns the public wire representation of this invitation.
func (i *PersistedInvitation) RenderDetail() Rendered[representation.InvitationDetail] {
	return render(representation.InvitationDetail{
		GroupName:   i.groupName,
		InviterName: i.inviterName,
		Status:      string(i.status),
		Role:        string(i.role),
		ExpiresAt:   i.expiresAt,
	})
}

// EnsureUsable fails with ErrGone when this invitation is revoked or has
// already expired at the given instant.
func (i *PersistedInvitation) EnsureUsable(now time.Time) error {
	return i.status.ensureUsable(now, i.expiresAt)
}

func (i *PersistedInvitation) Accept(ctx context.Context, user *PersistedUser, now time.Time, p Persister) error {
	if err := i.status.ensureAcceptable(now, i.expiresAt); err != nil {
		return err
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
		user, i.groupID,
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
		user.displayName, "", i.groupID,
	)
	if err != nil {
		return err
	}

	isPrimaryInt := 0
	_, err = p.Exec(
		ctx,
		`INSERT INTO user_participants (user_id, participant_id, role, is_primary)
		VALUES ($1, $2, $3, $4)`,
		user, participantID, string(i.role), isPrimaryInt,
	)
	if err != nil {
		return err
	}

	var updatedAt time.Time
	err = p.QueryRow(
		ctx,
		[]any{&updatedAt},
		`UPDATE group_invitations
		SET status = $1, accepted_by_user_id = $2, accepted_at = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING updated_at`,
		InvitationStatusAccepted, user, now, i.id,
	)
	if err != nil {
		return err
	}

	i.status = InvitationStatusAccepted
	i.acceptedByUserID = &user.id
	i.acceptedAt = &now
	i.updatedAt = updatedAt

	return nil
}

func (i *PersistedInvitation) Revoke(ctx context.Context, p Persister) error {
	if err := i.status.ensureRevokable(); err != nil {
		return err
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

func generateSecureToken(entropy io.Reader) (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(entropy, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
