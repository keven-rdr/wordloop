package domain

import "time"

type Lifetime struct {
	Absolute time.Duration
	Idle     time.Duration
}

func (l Lifetime) Expired(createdAt, lastUsedAt, now time.Time) bool {
	return !now.Before(createdAt.Add(l.Absolute)) || !now.Before(lastUsedAt.Add(l.Idle))
}

type SessionPolicy struct {
	User  Lifetime
	Admin Lifetime
}

func DefaultSessionPolicy() SessionPolicy {
	return SessionPolicy{
		User:  Lifetime{Absolute: 30 * 24 * time.Hour, Idle: 7 * 24 * time.Hour},
		Admin: Lifetime{Absolute: 8 * time.Hour, Idle: 30 * time.Minute},
	}
}

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

func (p SessionPolicy) For(r Role) Lifetime {
	if r == RoleAdmin {
		return p.Admin
	}
	return p.User
}
