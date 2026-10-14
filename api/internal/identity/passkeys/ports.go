package passkeys

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("sem convite valido nem sessao")
	ErrInvalid      = errors.New("cerimonia de passkey invalida")
	ErrNotFound     = errors.New("nao encontrado")
	ErrDuplicate    = errors.New("passkey ja cadastrada")
	ErrTooMany      = errors.New("limite de passkeys por conta atingido")
	ErrLastMethod   = errors.New("a conta ficaria sem forma de entrar")
)

const (
	MaxPerAccount = 10
	ChallengeTTL  = 5 * time.Minute
	MaxNameLength = 60
	DefaultName   = "Passkey"
)

type Purpose string

const (
	PurposeRegister Purpose = "register"
	PurposeLogin    Purpose = "login"
)

type Credential struct {
	ID             uuid.UUID
	CredentialID   []byte
	UserID         uuid.UUID
	PublicKey      []byte
	SignCount      uint32
	Transports     []string
	UserVerified   bool
	BackupEligible bool
	BackupState    bool
	CloneWarning   bool
	AAGUID         []byte
	Name           string
	CreatedAt      time.Time
	LastUsedAt     time.Time
}

type Owner struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
}

type Lookup func(credentialID []byte) (Credential, Owner, error)

type Ceremony interface {
	BeginRegistration(owner Owner, existing []Credential) (options, state []byte, err error)
	FinishRegistration(owner Owner, state, credentialJSON []byte) (Credential, error)
	BeginLogin() (options, state []byte, err error)
	FinishLogin(state, credentialJSON []byte, lookup Lookup) (Credential, Owner, error)
}

type Challenge struct {
	IDHash     []byte
	Purpose    Purpose
	UserID     uuid.UUID
	InviteHash []byte
	State      []byte
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

type Challenges interface {
	Save(ctx context.Context, c Challenge) error
	Take(ctx context.Context, idHash []byte, purpose Purpose, now time.Time) (Challenge, error)
}

type Store interface {
	List(ctx context.Context, userID uuid.UUID) ([]Credential, error)
	Find(ctx context.Context, credentialID []byte) (Credential, Owner, error)
	Owner(ctx context.Context, userID uuid.UUID) (Owner, error)
	PeekInvite(ctx context.Context, inviteHash []byte, now time.Time) (uuid.UUID, error)
	Add(ctx context.Context, c Credential) error
	AddWithInvite(ctx context.Context, c Credential, inviteHash []byte, now time.Time) error
	RecordLogin(ctx context.Context, c Credential, at time.Time) error
	Delete(ctx context.Context, userID, id uuid.UUID, allow func(passkeys, identities int) bool) error
}
