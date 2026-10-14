package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres/db"
	"github.com/keven-rdr/wordloop/api/internal/identity/passkeys"
)

type Passkeys struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

var (
	_ passkeys.Store      = (*Passkeys)(nil)
	_ passkeys.Challenges = (*Passkeys)(nil)
)

func NewPasskeys(pool *pgxpool.Pool) *Passkeys { return &Passkeys{pool: pool, q: db.New(pool)} }

func (r *Passkeys) List(ctx context.Context, userID uuid.UUID) ([]passkeys.Credential, error) {
	rows, err := r.q.ListPasskeysByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]passkeys.Credential, 0, len(rows))
	for _, row := range rows {
		out = append(out, passkeys.Credential{
			ID: row.ID, CredentialID: row.CredentialID, UserID: row.UserID, PublicKey: row.PublicKey, SignCount: signCount(row.SignCount),
			Transports: row.Transports, UserVerified: row.UserVerified, BackupEligible: row.BackupEligible, BackupState: row.BackupState,
			AAGUID: row.Aaguid, Name: row.Name, CreatedAt: row.CreatedAt.Time.UTC(), LastUsedAt: row.LastUsedAt.Time.UTC(),
		})
	}
	return out, nil
}

func (r *Passkeys) Find(ctx context.Context, credentialID []byte) (passkeys.Credential, passkeys.Owner, error) {
	row, err := r.q.GetPasskeyWithOwner(ctx, credentialID)
	if errors.Is(err, pgx.ErrNoRows) {
		return passkeys.Credential{}, passkeys.Owner{}, passkeys.ErrNotFound
	}
	if err != nil {
		return passkeys.Credential{}, passkeys.Owner{}, err
	}
	cred := passkeys.Credential{
		ID: row.ID, CredentialID: row.CredentialID, UserID: row.UserID, PublicKey: row.PublicKey, SignCount: signCount(row.SignCount),
		Transports: row.Transports, UserVerified: row.UserVerified, BackupEligible: row.BackupEligible, BackupState: row.BackupState,
		AAGUID: row.Aaguid, Name: row.Name, CreatedAt: row.CreatedAt.Time.UTC(), LastUsedAt: row.LastUsedAt.Time.UTC(),
	}
	return cred, passkeys.Owner{ID: row.UserID, Email: row.Email.String, DisplayName: row.DisplayName.String}, nil
}

func (r *Passkeys) Owner(ctx context.Context, userID uuid.UUID) (passkeys.Owner, error) {
	row, err := r.q.GetOwner(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return passkeys.Owner{}, passkeys.ErrNotFound
	}
	if err != nil {
		return passkeys.Owner{}, err
	}
	return passkeys.Owner{ID: row.ID, Email: row.Email.String, DisplayName: row.DisplayName.String}, nil
}

func (r *Passkeys) PeekInvite(ctx context.Context, inviteHash []byte, now time.Time) (uuid.UUID, error) {
	id, err := r.q.PeekInvite(ctx, db.PeekInviteParams{IDHash: inviteHash, ExpiresAt: ts(now)})
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, passkeys.ErrUnauthorized
	}
	return id, err
}

func (r *Passkeys) Add(ctx context.Context, c passkeys.Credential) error {
	return r.insert(ctx, r.q, c)
}

func (r *Passkeys) AddWithInvite(ctx context.Context, c passkeys.Credential, inviteHash []byte, now time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)
	owner, err := q.ConsumeInvite(ctx, db.ConsumeInviteParams{IDHash: inviteHash, UsedAt: ts(now)})
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && owner != c.UserID) {
		return passkeys.ErrUnauthorized
	}
	if err != nil {
		return fmt.Errorf("consumir o convite: %w", err)
	}
	if err := r.insert(ctx, q, c); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Passkeys) insert(ctx context.Context, q *db.Queries, c passkeys.Credential) error {
	n, err := q.CountPasskeysByUser(ctx, c.UserID)
	if err != nil {
		return err
	}
	if n >= passkeys.MaxPerAccount {
		return passkeys.ErrTooMany
	}
	err = q.InsertPasskey(ctx, db.InsertPasskeyParams{
		CredentialID: c.CredentialID, UserID: c.UserID, PublicKey: c.PublicKey, SignCount: int64(c.SignCount), Transports: nonNil(c.Transports),
		UserVerified: c.UserVerified, BackupEligible: c.BackupEligible, BackupState: c.BackupState, Aaguid: c.AAGUID,
		Name: c.Name, CreatedAt: ts(c.CreatedAt),
	})
	if isUniqueViolation(err) {
		return passkeys.ErrDuplicate
	}
	return err
}

func (r *Passkeys) RecordLogin(ctx context.Context, c passkeys.Credential, at time.Time) error {
	return r.q.RecordPasskeyLogin(ctx, db.RecordPasskeyLoginParams{
		CredentialID: c.CredentialID, SignCount: int64(c.SignCount), UserVerified: c.UserVerified, BackupState: c.BackupState, LastUsedAt: ts(at),
	})
}

func (r *Passkeys) Delete(ctx context.Context, userID, id uuid.UUID, allow func(passkeys, identities int) bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)
	if _, err := q.LockUser(ctx, userID); errors.Is(err, pgx.ErrNoRows) {
		return passkeys.ErrNotFound
	} else if err != nil {
		return err
	}
	creds, err := q.ListPasskeysByUser(ctx, userID)
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(creds, func(c db.PasskeyCredential) bool { return c.ID == id }) {
		return passkeys.ErrNotFound
	}
	identities, err := q.CountIdentitiesByUser(ctx, userID)
	if err != nil {
		return err
	}
	if !allow(len(creds), int(identities)) {
		return passkeys.ErrLastMethod
	}
	if _, err := q.DeletePasskey(ctx, db.DeletePasskeyParams{UserID: userID, ID: id}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Passkeys) Save(ctx context.Context, c passkeys.Challenge) error {
	if err := r.q.PurgeChallenges(ctx, ts(c.CreatedAt)); err != nil {
		return err
	}
	user := pgtype.UUID{}
	if c.UserID != uuid.Nil {
		user = pgtype.UUID{Bytes: c.UserID, Valid: true}
	}
	return r.q.SaveChallenge(ctx, db.SaveChallengeParams{
		IDHash: c.IDHash, Purpose: string(c.Purpose), UserID: user, InviteHash: c.InviteHash, State: c.State,
		CreatedAt: ts(c.CreatedAt), ExpiresAt: ts(c.ExpiresAt),
	})
}

func (r *Passkeys) Take(ctx context.Context, idHash []byte, purpose passkeys.Purpose, now time.Time) (passkeys.Challenge, error) {
	row, err := r.q.TakeChallenge(ctx, db.TakeChallengeParams{IDHash: idHash, Purpose: string(purpose), ExpiresAt: ts(now)})
	if errors.Is(err, pgx.ErrNoRows) {
		return passkeys.Challenge{}, passkeys.ErrNotFound
	}
	if err != nil {
		return passkeys.Challenge{}, err
	}
	return passkeys.Challenge{IDHash: idHash, Purpose: purpose, UserID: row.UserID.Bytes, InviteHash: row.InviteHash, State: row.State}, nil
}

func signCount(n int64) uint32 {
	switch {
	case n < 0:
		return 0
	case n > math.MaxUint32:
		return math.MaxUint32
	}
	return uint32(n)
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
