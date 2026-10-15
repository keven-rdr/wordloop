package domain

type TokenPurpose string

const (
	TokenInvite      TokenPurpose = "invite"
	TokenVerifyEmail TokenPurpose = "verify_email"
	TokenRecovery    TokenPurpose = "recovery"
)
