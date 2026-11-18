//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/platform/db"
	"github.com/keven-rdr/wordloop/api/internal/platform/db/dbtest"
	"github.com/keven-rdr/wordloop/api/internal/privacy/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/migrations"
)

// Segredos de mentira, plantados nas colunas sensiveis: se um deles aparecer na exportacao, o teste falha.
const (
	secretPasswordHash = "argon2id-SEGREDO-DA-SENHA"
	secretRefreshToken = "REFRESH-TOKEN-SEGREDO"
	secretPublicKey    = "CHAVE-PUBLICA-SEGREDO"
	secretSessionHash  = "HASH-DE-SESSAO-SEGREDO"
	secretTokenHash    = "HASH-DE-TOKEN-SEGREDO"
)

type fixture struct {
	pool  *pgxpool.Pool
	store *postgres.Store
}

func setup(t *testing.T) fixture {
	t.Helper()
	ctx := context.Background()
	url, _ := dbtest.Postgres(t)
	pool, err := db.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := db.Migrate(ctx, pool, migrations.FS); err != nil {
		t.Fatal(err)
	}
	return fixture{pool: pool, store: postgres.NewStore(pool)}
}

func (f fixture) exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := f.pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("%v: %s", err, sql)
	}
}

func (f fixture) count(t *testing.T, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := f.pool.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatalf("%v: %s", err, sql)
	}
	return n
}

// person cria uma pessoa com dados em todos os modulos: login, passkey, sessao, token, preferencias, meta, estudo,
// cartao e duas respostas no log (uma com correcao em review_void).
func (f fixture) person(t *testing.T, name, email string) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	var id uuid.UUID
	err := f.pool.QueryRow(ctx, `INSERT INTO app_user (display_name, email, email_verified_at, password_hash)
		VALUES ($1, $2, now(), $3) RETURNING id`, name, email, secretPasswordHash).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	f.exec(t, `INSERT INTO auth_identity (user_id, provider, subject) VALUES ($1, 'google', $2)`, id, "sub-"+name)
	f.exec(t, `INSERT INTO passkey_credential (user_id, credential_id, public_key, name, created_at)
		VALUES ($1, $2, $3, 'Meu celular', now())`, id, []byte("cred-"+name), []byte(secretPublicKey))
	f.exec(t, `INSERT INTO auth_session (id_hash, user_id, refresh_token_enc, expires_at, user_agent)
		VALUES ($1, $2, $3, now() + interval '1 day', 'Firefox')`, []byte(secretSessionHash+name), id, secretRefreshToken)
	f.exec(t, `INSERT INTO auth_token (id_hash, user_id, purpose, expires_at) VALUES ($1, $2, 'recovery', now() + interval '1 hour')`,
		[]byte(secretTokenHash+name), id)
	f.exec(t, `INSERT INTO user_settings (user_id, theme) VALUES ($1, 'dark')`, id)
	f.exec(t, `INSERT INTO goal_cycle (user_id, period_start, period_end, daily_minutes, days_target) VALUES ($1, '2026-11-10', '2026-11-16', 10, 5)`, id)
	f.exec(t, `INSERT INTO study_day (user_id, local_date, reviews, correct) VALUES ($1, '2026-11-12', 2, 1)`, id)
	f.studied(t, id)
	return id
}

// studied cria a sessao de estudo, o cartao e duas respostas no log, a primeira com uma correcao em review_void.
func (f fixture) studied(t *testing.T, user uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	var session, card, first uuid.UUID
	if err := f.pool.QueryRow(ctx, `INSERT INTO study_session (id, user_id, mode, planned_minutes, expires_at)
		VALUES (gen_random_uuid(), $1, 'mixed', 15, now() + interval '1 day') RETURNING id`, user).Scan(&session); err != nil {
		t.Fatal(err)
	}
	if err := f.pool.QueryRow(ctx, `INSERT INTO card (user_id, item_id, direction, state, stability, due_at)
		SELECT $1, id, 'receptive', 'review', 3, now() FROM item WHERE natural_key = 'word|house|noun|1' RETURNING id`, user).Scan(&card); err != nil {
		t.Fatal(err)
	}
	for i, correct := range []bool{false, true} {
		var review uuid.UUID
		err := f.pool.QueryRow(ctx, `INSERT INTO review_log (id, user_id, card_id, session_id, exercise_type, source, shown_at, answered_at,
				tz, day_cutoff_hour, local_date, response_ms, answer_raw, is_correct, grade, state_before, state_after, scheduler, scheduler_version)
			VALUES (gen_random_uuid(), $1, $2, $3, 'type', 'session', now() - interval '2 minutes', now() - interval '1 minute' + ($4 || ' seconds')::interval,
				'America/Sao_Paulo', 4, '2026-11-12', 3000, 'house', $5, 3, 'review', 'review', 'simple_v1', '1') RETURNING id`,
			user, card, session, i, correct).Scan(&review)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = review
		}
	}
	f.exec(t, `INSERT INTO review_void (id, review_id, reason) VALUES (gen_random_uuid(), $1, 'toque errado')`, first)
}
