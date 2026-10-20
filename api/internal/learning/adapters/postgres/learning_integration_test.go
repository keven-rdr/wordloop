//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/learning/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/learning/learningtest"
	"github.com/keven-rdr/wordloop/api/internal/platform/db"
	"github.com/keven-rdr/wordloop/api/internal/platform/db/dbtest"
	"github.com/keven-rdr/wordloop/api/internal/srs"
	"github.com/keven-rdr/wordloop/api/migrations"
)

var now = time.Date(2026, 10, 19, 15, 0, 0, 0, time.UTC)

// fixture sobe um Postgres real com as migracoes e o seed v0 (125 itens em rascunho), cria uma pessoa e publica
// alguns itens da unidade "casa e familia" para o aluno enxergar.
type fixture struct {
	pool   *pgxpool.Pool
	store  *postgres.Store
	user   uuid.UUID
	plan   app.PlanSession
	submit app.SubmitReviews
}

func setup(t *testing.T, admin bool) fixture {
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
	var user uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO app_user (display_name) VALUES ('Ana') RETURNING id`).Scan(&user); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE item SET status = 'published' WHERE natural_key IN
		('word|house|noun|1', 'word|hot|adj|1', 'word|cold|adj|1', 'word|big|adj|1', 'word|small|adj|1')`); err != nil {
		t.Fatal(err)
	}
	store := postgres.NewStore(pool)
	viewers := &learningtest.Viewers{Who: domain.Viewer{UserID: user, Admin: admin}}
	clock := func() time.Time { return now }
	sched := srs.SimpleV1{NoFuzz: true}
	return fixture{
		pool: pool, store: store, user: user,
		plan:   app.PlanSession{Viewers: viewers, Repo: store, Scheduler: sched, Now: clock, NewID: uuid.New},
		submit: app.SubmitReviews{Viewers: viewers, Repo: store, Scheduler: sched, Now: clock},
	}
}

func count(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestPlanCreatesReceptiveCardsOnlyForPublishedItemsAndIsStable(t *testing.T) {
	f := setup(t, false)
	ctx := context.Background()
	pack, err := f.plan.Run(ctx, app.PlanRequest{Minutes: 30, TZ: "Europe/Lisbon"})
	if err != nil {
		t.Fatal(err)
	}
	if len(pack.Cards) != 5 || pack.NewLimit != 6 {
		t.Fatalf("aluno: so os 5 itens publicados (30 min = 6 novas): %d cartoes, limite %d", len(pack.Cards), pack.NewLimit)
	}
	for _, c := range pack.Cards {
		if c.Content.Term == "" || c.Content.Gloss == "" || c.Card.Direction != domain.Receptive || len(c.Options) == 0 {
			t.Errorf("carta sem conteudo ou alternativas: %+v", c)
		}
		correct := 0
		for _, o := range c.Options {
			if o.ItemID == c.Card.ItemID {
				correct++
			} else if o.Gloss == c.Content.Gloss {
				t.Errorf("distrator com a mesma traducao da resposta: %+v", o)
			}
		}
		if correct != 1 {
			t.Errorf("a resposta certa aparece uma vez: %+v", c.Options)
		}
	}
	if n := count(t, f.pool, `SELECT count(*) FROM card WHERE user_id = $1`, f.user); n != 5 {
		t.Errorf("cartoes no banco: %d", n)
	}
	if n := count(t, f.pool, `SELECT count(*) FROM user_settings WHERE user_id = $1 AND tz = 'Europe/Lisbon'`, f.user); n != 1 {
		t.Error("o fuso informado fica guardado")
	}
	again, err := f.plan.Run(ctx, app.PlanRequest{Minutes: 30})
	if err != nil || len(again.Cards) != 5 || count(t, f.pool, `SELECT count(*) FROM card`) != 5 {
		t.Fatalf("o plano repetido reaproveita os cartoes nao respondidos: %d %v", len(again.Cards), err)
	}
}

func TestAdminAlsoGetsDraftItems(t *testing.T) {
	f := setup(t, true)
	pack, err := f.plan.Run(context.Background(), app.PlanRequest{Minutes: 60})
	if err != nil || len(pack.Cards) != 8 {
		t.Fatalf("admin: ate o teto de 8 novas, incluindo rascunhos: %d %v", len(pack.Cards), err)
	}
}

func TestSubmitReviewsWritesTheImmutableLogAndIsIdempotent(t *testing.T) {
	f := setup(t, false)
	ctx := context.Background()
	pack, _ := f.plan.Run(ctx, app.PlanRequest{Minutes: 15})
	card := pack.Cards[0].Card
	a := domain.Answer{
		ID: uuid.New(), CardID: card.ID, SessionID: pack.SessionID, Exercise: domain.Choose, ChosenItemID: card.ItemID,
		ShownAt: now.Add(-time.Minute), AnsweredAt: now.Add(-50 * time.Second), ResponseMS: 3200,
		Difficulty: srs.DifficultyMedium, ClientSeq: 7,
	}
	res, err := f.submit.Run(ctx, []domain.Answer{a})
	if err != nil || res[0].Status != domain.Applied || res[0].Card.State.Stability != 2 {
		t.Fatalf("1a resposta: %+v %v", res, err)
	}
	var stability float64
	var state string
	if err := f.pool.QueryRow(ctx, `SELECT stability, state FROM card WHERE id = $1`, card.ID).Scan(&stability, &state); err != nil || stability != 2 || state != "review" {
		t.Fatalf("cartao salvo: %v %v %v", stability, state, err)
	}
	var tz string
	var local time.Time
	var isCorrect bool
	if err := f.pool.QueryRow(ctx, `SELECT tz, local_date, is_correct FROM review_log WHERE id = $1`, a.ID).Scan(&tz, &local, &isCorrect); err != nil ||
		tz != "America/Sao_Paulo" || local.Format("2006-01-02") != "2026-10-19" || !isCorrect {
		t.Fatalf("linha do log: %v %v %v %v", tz, local, isCorrect, err)
	}

	res, err = f.submit.Run(ctx, []domain.Answer{a})
	if err != nil || res[0].Status != domain.Duplicate || count(t, f.pool, `SELECT count(*) FROM review_log`) != 1 {
		t.Fatalf("reenvio nao muda nada: %+v %v", res, err)
	}
	for _, sql := range []string{`UPDATE review_log SET grade = 1`, `DELETE FROM review_log`} {
		if _, err := f.pool.Exec(ctx, sql); err == nil {
			t.Errorf("o log e append-only, mas aceitou: %s", sql)
		}
	}
}

func TestOutOfOrderAnswerReplaysTheLogFromTheDatabase(t *testing.T) {
	f := setup(t, false)
	ctx := context.Background()
	pack, _ := f.plan.Run(ctx, app.PlanRequest{Minutes: 15})
	card := pack.Cards[0].Card
	day := func(n int) time.Time { return now.Add(-10*24*time.Hour + time.Duration(n)*24*time.Hour) }
	answer := func(at time.Time) domain.Answer {
		return domain.Answer{ID: uuid.New(), CardID: card.ID, Exercise: domain.Choose, ChosenItemID: card.ItemID, ShownAt: at.Add(-time.Second), AnsweredAt: at, Difficulty: srs.DifficultyMedium}
	}
	first, second, third := answer(day(0)), answer(day(3)), answer(day(6))
	for _, a := range []domain.Answer{first, third, second} {
		if res, err := f.submit.Run(ctx, []domain.Answer{a}); err != nil || res[0].Status != domain.Applied {
			t.Fatalf("resposta: %+v %v", res, err)
		}
	}
	want, _ := srs.Replay(srs.SimpleV1{NoFuzz: true}, card.ID.String(), []srs.Answer{
		{Review: srs.Review{Correct: true, Difficulty: srs.DifficultyMedium}, At: first.AnsweredAt},
		{Review: srs.Review{Correct: true, Difficulty: srs.DifficultyMedium}, At: second.AnsweredAt},
		{Review: srs.Review{Correct: true, Difficulty: srs.DifficultyMedium}, At: third.AnsweredAt},
	})
	var stability float64
	var reps int
	if err := f.pool.QueryRow(ctx, `SELECT stability, reps FROM card WHERE id = $1`, card.ID).Scan(&stability, &reps); err != nil ||
		stability != want.Stability || reps != want.Reps {
		t.Fatalf("estado = replay do log: S=%v reps=%d, esperava S=%v reps=%d (%v)", stability, reps, want.Stability, want.Reps, err)
	}
}

func TestProductiveCardAppearsAfterTheReceptiveGetsStrong(t *testing.T) {
	f := setup(t, false)
	ctx := context.Background()
	pack, _ := f.plan.Run(ctx, app.PlanRequest{Minutes: 15})
	card := pack.Cards[0].Card
	// forca um receptivo forte e revisado ha 3 dias
	if _, err := f.pool.Exec(ctx, `UPDATE card SET state = 'review', stability = 10, reps = 4, last_review_at = $2, due_at = $3 WHERE id = $1`,
		card.ID, now.Add(-72*time.Hour), now.Add(7*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	next, err := f.plan.Run(ctx, app.PlanRequest{Minutes: 15, Mode: domain.ModeLearn})
	if err != nil {
		t.Fatal(err)
	}
	productive := 0
	for _, c := range next.Cards {
		if c.Card.Direction == domain.Productive && c.Card.ItemID == card.ItemID {
			productive++
			if len(c.Options) != 0 || c.Card.Exercise() != domain.Type {
				t.Errorf("produtivo digita, sem alternativas: %+v", c)
			}
		}
	}
	if productive != 1 {
		t.Fatalf("o cartao produtivo nasce uma vez: %d", productive)
	}
	// digitar com um erro de uma letra vale Hard sem lapso
	var prod domain.Card
	for _, c := range next.Cards {
		if c.Card.Direction == domain.Productive {
			prod = c.Card
		}
	}
	typed := domain.Answer{ID: uuid.New(), CardID: prod.ID, Exercise: domain.Type, Text: "huose", ShownAt: now.Add(-time.Minute), AnsweredAt: now.Add(-30 * time.Second)}
	res, err := f.submit.Run(ctx, []domain.Answer{typed})
	if err != nil || res[0].Status != domain.Applied || res[0].Grade != srs.Hard || !res[0].Typo || res[0].Card.State.Lapses != 0 {
		t.Fatalf("erro de digitacao: %+v %v", res, err)
	}
}

func TestAnotherUsersCardIsNotFound(t *testing.T) {
	f := setup(t, false)
	ctx := context.Background()
	pack, _ := f.plan.Run(ctx, app.PlanRequest{Minutes: 15})
	var other uuid.UUID
	if err := f.pool.QueryRow(ctx, `INSERT INTO app_user (display_name) VALUES ('Outra') RETURNING id`).Scan(&other); err != nil {
		t.Fatal(err)
	}
	viewers := &learningtest.Viewers{Who: domain.Viewer{UserID: other}}
	intruder := app.SubmitReviews{Viewers: viewers, Repo: f.store, Scheduler: srs.SimpleV1{}, Now: func() time.Time { return now }}
	card := pack.Cards[0].Card
	res, err := intruder.Run(ctx, []domain.Answer{{
		ID: uuid.New(), CardID: card.ID, Exercise: domain.Choose, ChosenItemID: card.ItemID,
		ShownAt: now.Add(-time.Minute), AnsweredAt: now.Add(-30 * time.Second),
	}})
	if err != nil || res[0].Status != domain.Rejected || res[0].Reason != domain.ReasonCardNotFound {
		t.Fatalf("cartao de outra pessoa: %+v %v", res, err)
	}
}
