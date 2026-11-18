//go:build integration

package postgres_test

import (
	"context"
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/platform/db"
	"github.com/keven-rdr/wordloop/api/internal/platform/db/dbtest"
	"github.com/keven-rdr/wordloop/api/internal/progress/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
	"github.com/keven-rdr/wordloop/api/migrations"
)

type fixture struct {
	pool  *pgxpool.Pool
	store *postgres.Store
	user  uuid.UUID
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
	var user uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO app_user (display_name) VALUES ('Ana') RETURNING id`).Scan(&user); err != nil {
		t.Fatal(err)
	}
	return fixture{pool: pool, store: postgres.NewStore(pool), user: user}
}

func (f fixture) exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := f.pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("%v: %s", err, sql)
	}
}

// card cria o cartao receptivo de um item do seed.
func (f fixture) card(t *testing.T, user uuid.UUID, key, state string, stability float64, due time.Time, suspended bool) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := f.pool.QueryRow(context.Background(), `
		INSERT INTO card (user_id, item_id, direction, state, stability, due_at, suspended)
		SELECT $1, id, 'receptive', $3, $4, $5, $6 FROM item WHERE natural_key = $2 RETURNING id`,
		user, key, state, stability, due, suspended).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// log grava uma resposta no log (so as colunas obrigatorias).
func (f fixture) log(t *testing.T, user, card uuid.UUID, date string, correct, typo bool) {
	t.Helper()
	grade := 1 // Again; o acerto vale Good
	if correct {
		grade = 3
	}
	f.exec(t, `
		INSERT INTO review_log (id, user_id, card_id, exercise_type, source, shown_at, answered_at, tz, day_cutoff_hour,
			local_date, response_ms, is_correct, typo_flag, grade, state_before, state_after, scheduler, scheduler_version)
		VALUES (gen_random_uuid(), $1, $2, 'type', 'session', $4::timestamptz - interval '5 seconds', $4::timestamptz,
			'America/Sao_Paulo', 4, $3::date, 4000, $5, $6, $7, 'review', 'review', 'simple_v1', '1')`,
		user, card, date, date+"T15:00:00Z", correct, typo, grade)
}

func TestSettingsFallBackToTheDefaultsAndReadTheSavedOnes(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	got, err := f.store.Settings(ctx, f.user)
	if err != nil || got != domain.DefaultSettings {
		t.Fatalf("sem ajustes salvos: %+v %v", got, err)
	}
	f.exec(t, `INSERT INTO user_settings (user_id, tz, day_cutoff_hour) VALUES ($1, 'Europe/Lisbon', 3)`, f.user)
	got, err = f.store.Settings(ctx, f.user)
	if err != nil || got != (domain.Settings{TZ: "Europe/Lisbon", CutoffHour: 3}) {
		t.Fatalf("ajustes salvos: %+v %v", got, err)
	}
}

func TestDaysListsOnlyStudiedDaysOfThePersonInsideTheRange(t *testing.T) {
	f := setup(t)
	var other uuid.UUID
	if err := f.pool.QueryRow(context.Background(), `INSERT INTO app_user (display_name) VALUES ('Bia') RETURNING id`).Scan(&other); err != nil {
		t.Fatal(err)
	}
	f.exec(t, `INSERT INTO study_day (user_id, local_date, reviews, correct, new_items, active_ms, known_lexemes) VALUES
		($1, '2026-10-17', 5, 4, 2, 60000, 3), ($1, '2026-10-18', 0, 0, 0, 0, NULL), ($1, '2026-10-19', 8, 8, 0, 90000, NULL),
		($1, '2026-09-01', 9, 9, 0, 1000, NULL), ($2, '2026-10-19', 99, 99, 0, 1, NULL)`, f.user, other)

	days, err := f.store.Days(context.Background(), f.user, day("2026-10-01"), day("2026-10-19"))
	if err != nil || len(days) != 2 {
		t.Fatalf("2 dias com estudo no intervalo: %+v %v", days, err)
	}
	if days[0].Date != day("2026-10-17") || days[0].Reviews != 5 || days[0].NewItems != 2 || days[0].ActiveMS != 60000 || days[0].KnownWords != 3 || days[1].KnownWords != 0 {
		t.Errorf("dia 17: %+v", days[0])
	}
}

func TestForecastCountsByDayInThePersonsTimeZoneAndFoldsOverdueIntoToday(t *testing.T) {
	f := setup(t)
	due := func(s string) time.Time { return day(s).Add(12 * time.Hour) }         // 09:00 em Sao Paulo
	f.card(t, f.user, "word|house|noun|1", "review", 3, due("2026-10-10"), false) // atrasada: hoje
	f.card(t, f.user, "word|hot|adj|1", "review", 3, due("2026-10-20"), false)    // amanha
	f.card(t, f.user, "word|cold|adj|1", "learning", 1, due("2026-10-20"), false) // amanha
	f.card(t, f.user, "word|big|adj|1", "review", 3, due("2026-11-30"), false)    // alem da previsao
	f.card(t, f.user, "word|small|adj|1", "review", 3, due("2026-10-21"), true)   // suspensa
	// 02:00 em Sao Paulo (05:00 UTC) do dia 20 ainda e o dia 19 por causa do corte de 04:00: conta em hoje.
	f.exec(t, `UPDATE card SET due_at = '2026-10-20T05:00:00Z' WHERE item_id = (SELECT id FROM item WHERE natural_key = 'word|cold|adj|1')`)

	got, err := f.store.Forecast(context.Background(), f.user, domain.DefaultSettings, day("2026-10-19"), 7)
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != 2 || got[1] != 1 || len(got) != 2 {
		t.Errorf("hoje: atrasada + madrugada; amanha: 1; o resto fica de fora: %v", got)
	}
}

func TestKnownWordsCountsReceptiveCardsInReviewWithStabilityOfSevenDays(t *testing.T) {
	f := setup(t)
	soon := day("2026-10-20")
	f.card(t, f.user, "word|house|noun|1", "review", 7, soon, false)
	f.card(t, f.user, "word|hot|adj|1", "review", 6.9, soon, false)  // ainda nao
	f.card(t, f.user, "word|cold|adj|1", "learning", 9, soon, false) // nao esta em revisao
	n, err := f.store.KnownWords(context.Background(), f.user)
	if err != nil || n != 1 {
		t.Errorf("so a que passou de 7 dias em revisao conta: %d %v", n, err)
	}
}

func TestMostMissedRanksTheWordsTheyMissTheMostAndIgnoresTypos(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	soon := day("2026-10-20")
	house := f.card(t, f.user, "word|house|noun|1", "review", 3, soon, false)
	hot := f.card(t, f.user, "word|hot|adj|1", "review", 3, soon, false)
	cold := f.card(t, f.user, "word|cold|adj|1", "review", 3, soon, false)
	for i := 0; i < 3; i++ {
		f.log(t, f.user, house, "2026-10-18", false, false)
	}
	f.log(t, f.user, hot, "2026-10-17", false, false)
	f.log(t, f.user, hot, "2026-10-19", false, false)
	f.log(t, f.user, hot, "2026-10-19", true, false)   // acerto nao conta
	f.log(t, f.user, cold, "2026-10-19", false, true)  // erro de 1 letra nao conta
	f.log(t, f.user, cold, "2026-08-01", false, false) // antigo demais

	got, err := f.store.MostMissed(ctx, f.user, day("2026-09-20"), "pt-BR", 5)
	if err != nil || len(got) != 2 {
		t.Fatalf("2 palavras com erro: %+v %v", got, err)
	}
	if got[0].Term != "house" || got[0].Misses != 3 || got[0].Gloss == "" || got[1].Term != "hot" || got[1].Misses != 2 {
		t.Errorf("ordem e contagem: %+v", got)
	}
	limited, err := f.store.MostMissed(ctx, f.user, day("2026-09-20"), "pt-BR", 1)
	if err != nil || len(limited) != 1 {
		t.Errorf("limite: %+v %v", limited, err)
	}
	var stranger uuid.UUID
	if err := f.pool.QueryRow(ctx, `INSERT INTO app_user (display_name) VALUES ('Caio') RETURNING id`).Scan(&stranger); err != nil {
		t.Fatal(err)
	}
	if none, err := f.store.MostMissed(ctx, stranger, day("2026-09-20"), "pt-BR", 5); err != nil || len(none) != 0 {
		t.Errorf("os erros de uma pessoa nao aparecem para outra: %+v %v", none, err)
	}
}

// O preenchimento retroativo da migracao 00015 refaz study_day a partir do log; roda o proprio SQL da migracao.
func TestStudyDayBackfillRebuildsTheDayFromTheLog(t *testing.T) {
	f := setup(t)
	soon := day("2026-10-20")
	house := f.card(t, f.user, "word|house|noun|1", "review", 3, soon, false)
	f.log(t, f.user, house, "2026-10-18", false, false) // lapso: nota Again num cartao ja em revisao
	f.log(t, f.user, house, "2026-10-18", true, false)
	f.exec(t, `INSERT INTO review_log (id, user_id, card_id, exercise_type, source, shown_at, answered_at, tz, day_cutoff_hour,
			local_date, response_ms, is_correct, typo_flag, grade, state_before, state_after, scheduler, scheduler_version)
		VALUES (gen_random_uuid(), $1, $2, 'choose', 'session', now(), now(), 'America/Sao_Paulo', 4, '2026-10-18', 1000,
			true, false, 3, 'new', 'review', 'simple_v1', '1')`, f.user, house)
	f.exec(t, `DELETE FROM study_day`)

	body, err := fs.ReadFile(migrations.FS, "00015_study_day.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	start := strings.Index(sql, "INSERT INTO study_day")
	end := start + strings.Index(sql[start:], ";") + 1
	f.exec(t, sql[start:end])

	var reviews, correct, lapses, fresh int
	var ms int64
	err = f.pool.QueryRow(context.Background(), `SELECT reviews, correct, lapses, new_items, active_ms FROM study_day WHERE user_id = $1 AND local_date = '2026-10-18'`, f.user).
		Scan(&reviews, &correct, &lapses, &fresh, &ms)
	if err != nil || reviews != 3 || correct != 2 || lapses != 1 || fresh != 1 || ms != 9000 {
		t.Errorf("dia refeito: reviews=%d correct=%d lapses=%d new=%d ms=%d (%v)", reviews, correct, lapses, fresh, ms, err)
	}
}
