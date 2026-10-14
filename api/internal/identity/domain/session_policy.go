package domain

import "time"

// Lifetime e a duracao de uma sessao: ate Absolute depois do login, e no maximo Idle sem uso.
type Lifetime struct {
	Absolute time.Duration
	Idle     time.Duration
}

// Expired diz se a sessao nao vale mais em now.
func (l Lifetime) Expired(createdAt, lastUsedAt, now time.Time) bool {
	return !now.Before(createdAt.Add(l.Absolute)) || !now.Before(lastUsedAt.Add(l.Idle))
}

// SessionPolicy define os prazos por papel. Admin tem sessao curta: roubar o cookie dele rende pouco tempo.
type SessionPolicy struct {
	User  Lifetime
	Admin Lifetime
}

// DefaultSessionPolicy: aluno 30 dias (7 sem uso); admin 8 horas (30 minutos sem uso).
func DefaultSessionPolicy() SessionPolicy {
	return SessionPolicy{
		User:  Lifetime{Absolute: 30 * 24 * time.Hour, Idle: 7 * 24 * time.Hour},
		Admin: Lifetime{Absolute: 8 * time.Hour, Idle: 30 * time.Minute},
	}
}

// OrDefault completa com o padrao qualquer prazo deixado em zero: um Idle ou Absolute esquecido nunca vira
// "expira na hora" (que desconectaria todo mundo).
func (p SessionPolicy) OrDefault() SessionPolicy {
	d := DefaultSessionPolicy()
	fill := func(l *Lifetime, def Lifetime) {
		if l.Absolute <= 0 {
			l.Absolute = def.Absolute
		}
		if l.Idle <= 0 {
			l.Idle = def.Idle
		}
	}
	fill(&p.User, d.User)
	fill(&p.Admin, d.Admin)
	return p
}

// For devolve o prazo do papel; papel desconhecido recebe o do usuario comum.
func (p SessionPolicy) For(r Role) Lifetime {
	if r == RoleAdmin {
		return p.Admin
	}
	return p.User
}
