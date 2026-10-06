//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"example.com/wordloop/api/internal/learning/adapters/postgres/db"
)

const seedPath = "../../../../../../fase-6/seed/out/sample-seed.sql"

func TestReviewFlow(t *testing.T) {
	ctx := context.Background()
	ctr, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("wl"), postgres.WithUsername("wl"), postgres.WithPassword("wl"),
		postgres.BasicWaitStrategies())
	if err != nil {
		t.Fatalf("subir contêiner: %v", err)
	}
	t.Cleanup(func() { _ = ctr.Terminate(ctx) })

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	// 1) migrações com goose (provider)
	prov, err := goose.NewProvider(goose.DialectPostgres, sqlDB, os.DirFS("../../../../migrations"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := prov.Up(ctx)
	if err != nil {
		t.Fatalf("goose up: %v", err)
	}
	t.Logf("goose aplicou %d migrações", len(res))
	if len(res) != 2 {
		t.Fatalf("esperava 2 migrações, veio %d", len(res))
	}

	// 2) seed da amostra por cima do schema migrado
	seed, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.ExecContext(ctx, string(seed)); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// 3) consultas geradas pelo sqlc, com pgx
	pool, err := openPgx(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := db.New(pool)

	var courseID, itemID uuid.UUID
	if err := pool.QueryRow(ctx, "select course_id, id from item where natural_key='word|cat|noun|1'").Scan(&courseID, &itemID); err != nil {
		t.Fatal(err)
	}
	it, err := q.GetItemByNaturalKey(ctx, db.GetItemByNaturalKeyParams{CourseID: courseID, NaturalKey: "word|cat|noun|1"})
	if err != nil || it.Kind != "word" {
		t.Fatalf("GetItemByNaturalKey: %+v %v", it, err)
	}

	user, card := uuid.New(), uuid.New()
	mustExec(t, pool, "insert into app_user(id) values ($1)", user)
	mustExec(t, pool, `insert into card(id,user_id,item_id,direction,state,stability,due_at,scheduler)
	                   values ($1,$2,$3,'receptive','review',9,now()-interval '1 day','simple_v1')`, card, user, itemID)

	due, err := q.ListDueCards(ctx, db.ListDueCardsParams{UserID: user, DueAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}, Limit: 10})
	if err != nil || len(due) != 1 {
		t.Fatalf("ListDueCards: %d %v", len(due), err)
	}

	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	rid := uuid.New()
	p := db.InsertReviewParams{ID: rid, UserID: user, CardID: card, ExerciseType: "choice_en_pt", Source: "session",
		ShownAt: now, AnsweredAt: now, Tz: "America/Sao_Paulo", DayCutoffHour: 4,
		LocalDate: pgtype.Date{Time: time.Now(), Valid: true}, IsCorrect: true, Grade: 3,
		StateBefore: "review", StateAfter: "review", Scheduler: "simple_v1", SchedulerVersion: "1"}
	for i, want := range []int64{1, 0} { // 2º envio do mesmo UUID não insere
		n, err := q.InsertReview(ctx, p)
		if err != nil || n != want {
			t.Fatalf("InsertReview #%d: n=%d err=%v (esperava %d)", i+1, n, err, want)
		}
	}
	if _, err := pool.Exec(ctx, "update review_log set grade=1 where id=$1", rid); err == nil {
		t.Fatal("UPDATE no review_log deveria falhar (append-only)")
	}
	k, err := q.CountKnownLexemes(ctx, user)
	if err != nil || k != 1 {
		t.Fatalf("CountKnownLexemes: %d %v", k, err)
	}
}
