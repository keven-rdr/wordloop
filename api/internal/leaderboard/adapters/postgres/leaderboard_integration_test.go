//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/leaderboard/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/leaderboard/domain"
	"github.com/keven-rdr/wordloop/api/internal/platform/db"
	"github.com/keven-rdr/wordloop/api/internal/platform/db/dbtest"
	"github.com/keven-rdr/wordloop/api/migrations"
)

type fixture struct {
	pool  *pgxpool.Pool
	store *postgres.Store
}

func day(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
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

// person cria alguem com pontos num dia (e o tempo ativo, em minutos) e, se nickname != "", o perfil.
func (f fixture) person(t *testing.T, nickname string, visible bool) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := f.pool.QueryRow(context.Background(), `INSERT INTO app_user (display_name) VALUES ('Nome Real') RETURNING id`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if nickname != "" {
		f.exec(t, `INSERT INTO leaderboard_profile (user_id, nickname, visible) VALUES ($1, $2, $3)`, id, nickname, visible)
	}
	return id
}

func (f fixture) studied(t *testing.T, user uuid.UUID, date string, points, minutes int) {
	t.Helper()
	f.exec(t, `INSERT INTO study_day (user_id, local_date, reviews, points, active_ms) VALUES ($1, $2, 1, $3, $4)`,
		user, date, points, int64(minutes)*60000)
}

var weekFrom, weekTo = day("2026-11-23"), day("2026-11-29")

func TestRankingCountsOnlyVisiblePeopleSharesTiesAndShowsMeAnyway(t *testing.T) {
	f := setup(t)
	ana, bia, caio := f.person(t, "ana", true), f.person(t, "bia", true), f.person(t, "caio", false)
	me := f.person(t, "", false)
	f.studied(t, ana, "2026-11-24", 100, 1)
	f.studied(t, bia, "2026-11-25", 100, 1)  // empate com ana
	f.studied(t, caio, "2026-11-25", 500, 1) // nao participa: nao aparece nem conta
	f.studied(t, me, "2026-11-26", 40, 1)
	f.studied(t, ana, "2026-11-16", 999, 1) // semana anterior: fora da semana

	rows, err := f.store.Ranking(context.Background(), me, weekFrom, weekTo, 10)
	if err != nil || len(rows) != 3 {
		t.Fatalf("ana, bia e a minha linha: %+v %v", rows, err)
	}
	if rows[0].Position != 1 || rows[1].Position != 1 || rows[0].Nickname != "ana" || rows[1].Nickname != "bia" {
		t.Errorf("empate divide a 1a posicao, desempate pelo apelido: %+v", rows)
	}
	if rows[2].UserID != me || rows[2].Position != 3 || rows[2].Nickname != "" || rows[2].Points != 40 {
		t.Errorf("a minha linha sem participar: posicao so para mim: %+v", rows[2])
	}
	all, _ := f.store.Ranking(context.Background(), me, day("0001-01-01"), day("9999-12-31"), 10)
	if all[0].Nickname != "ana" || all[0].Points != 1099 {
		t.Errorf("geral soma todas as semanas: %+v", all)
	}
}

func TestRankingAddsTheGoalBonuses(t *testing.T) {
	f := setup(t)
	ana := f.person(t, "ana", true)
	f.exec(t, `INSERT INTO goal_cycle (user_id, period_start, period_end, daily_minutes, days_target) VALUES ($1, $2, $3, 10, 7)`,
		ana, weekFrom, weekTo)
	for i := 0; i < 7; i++ {
		f.studied(t, ana, weekFrom.AddDate(0, 0, i).Format("2006-01-02"), 10, 10) // 7 dias com os 10 minutos da meta
	}
	rows, err := f.store.Ranking(context.Background(), ana, weekFrom, weekTo, 10)
	if err != nil || len(rows) != 1 || rows[0].Points != 7*10+7*50+300 {
		t.Errorf("70 das respostas + 7 x 50 + 300 da semana cheia: %+v %v", rows, err)
	}
	if n, _ := f.store.DaysMet(context.Background(), ana, weekFrom, weekTo); n != 7 {
		t.Errorf("dias da meta cumpridos: %d", n)
	}
}

func TestProfileIsUniqueIgnoringCaseAndGoesAwayWithTheAccount(t *testing.T) {
	f := setup(t)
	ana, bia := f.person(t, "", false), f.person(t, "", false)
	if _, err := f.store.SaveProfile(context.Background(), ana, domain.Profile{Nickname: "Ana", Visible: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.store.SaveProfile(context.Background(), bia, domain.Profile{Nickname: "ANA"}); !errors.Is(err, domain.ErrNicknameTaken) {
		t.Errorf("apelido repetido sem diferenciar maiusculas: %v", err)
	}
	if p, found, err := f.store.Profile(context.Background(), ana); err != nil || !found || p.Nickname != "Ana" || !p.Visible {
		t.Errorf("perfil: %+v %v %v", p, found, err)
	}
	f.exec(t, `DELETE FROM app_user WHERE id = $1`, ana)
	if _, found, _ := f.store.Profile(context.Background(), ana); found {
		t.Error("o perfil sai junto com a conta")
	}
}
