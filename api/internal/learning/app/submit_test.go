package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/learning/learningtest"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

func submitter(s *learningtest.Store, at time.Time) app.SubmitReviews {
	return app.SubmitReviews{
		Viewers: &learningtest.Viewers{Who: student}, Repo: s, Scheduler: srs.SimpleV1{NoFuzz: true},
		Now: func() time.Time { return at },
	}
}

// newCard cria um cartao receptivo ainda nao respondido para o 1o item.
func newCard(s *learningtest.Store, dir domain.Direction) *domain.Card {
	id := uuid.New()
	c := &domain.Card{ID: id, ItemID: s.Items[0].ID, Direction: dir, Scheduler: "simple_v1", State: srs.CardState{ID: id.String(), State: srs.StateNew}}
	s.Cards[id] = c
	return c
}

func choose(c *domain.Card, picked uuid.UUID, answeredAt time.Time) domain.Answer {
	return domain.Answer{
		ID: uuid.New(), CardID: c.ID, Exercise: domain.Choose, ChosenItemID: picked,
		ShownAt: answeredAt.Add(-5 * time.Second), AnsweredAt: answeredAt, Difficulty: srs.DifficultyMedium,
	}
}

func typed(c *domain.Card, text string, answeredAt time.Time) domain.Answer {
	return domain.Answer{
		ID: uuid.New(), CardID: c.ID, Exercise: domain.Type, Text: text,
		ShownAt: answeredAt.Add(-5 * time.Second), AnsweredAt: answeredAt, Difficulty: srs.DifficultyMedium,
	}
}

func one(t *testing.T, u app.SubmitReviews, a domain.Answer) domain.Result {
	t.Helper()
	res, err := u.Run(context.Background(), []domain.Answer{a})
	if err != nil || len(res) != 1 {
		t.Fatalf("envio: %v %v", res, err)
	}
	return res[0]
}

func TestSubmitCorrectChoiceUpdatesTheCardAndWritesTheLog(t *testing.T) {
	s := store("house", "cat")
	c := newCard(s, domain.Receptive)
	a := choose(c, c.ItemID, now.Add(-time.Minute))
	res := one(t, submitter(s, now), a)
	if res.Status != domain.Applied || !res.Correct || res.Grade != srs.Good || res.Card == nil {
		t.Fatalf("resultado: %+v", res)
	}
	got := s.Cards[c.ID].State
	if got.State != srs.StateReview || got.Stability != 2 || got.Reps != 1 || !got.DueAt.Equal(a.AnsweredAt.Add(48*time.Hour)) {
		t.Errorf("1o acerto medio: S=2, vence em 2 dias: %+v", got)
	}
	if len(s.Log) != 1 {
		t.Fatalf("log: %d linhas", len(s.Log))
	}
	row := s.Log[0]
	if !row.Correct || row.Grade != srs.Good || row.Before.State != srs.StateNew || row.After.State != srs.StateReview ||
		row.TZ != "America/Sao_Paulo" || row.LocalDate.Format("2006-01-02") != "2026-10-19" || row.UserID != user || row.Scheduler != "simple_v1" {
		t.Errorf("linha do log: %+v", row)
	}
}

func TestSubmitWrongChoiceIsAgainButNotALapseOnTheFirstContact(t *testing.T) {
	s := store("house", "cat")
	c := newCard(s, domain.Receptive)
	res := one(t, submitter(s, now), choose(c, s.Items[1].ID, now.Add(-time.Minute)))
	if res.Status != domain.Applied || res.Correct || res.Grade != srs.Again {
		t.Fatalf("errou: %+v", res)
	}
	if st := s.Cards[c.ID].State; st.State != srs.StateLearning || st.Lapses != 0 {
		t.Errorf("1o erro vai para aprendizado, sem lapso: %+v", st)
	}
}

func TestSubmitTypedAnswerIsCorrectedOnTheServer(t *testing.T) {
	s := store("house", "cat")
	cases := map[string]struct {
		text    string
		correct bool
		typo    bool
		grade   srs.Grade
	}{
		"certo":     {"House", true, false, srs.Good},
		"digitacao": {"hause", false, true, srs.Hard},
		"errado":    {"mouse x", false, false, srs.Again},
		"outra":     {"cat", false, false, srs.Again},
	}
	for name, want := range cases {
		c := newCard(s, domain.Productive)
		res := one(t, submitter(s, now), typed(c, want.text, now.Add(-time.Minute)))
		if res.Status != domain.Applied || res.Correct != want.correct || res.Typo != want.typo || res.Grade != want.grade {
			t.Errorf("%s: %+v", name, res)
		}
	}
}

func TestSubmitIgnoresWhatTheClientSaysAboutTheAnswerButLogsIt(t *testing.T) {
	s := store("house", "cat")
	c := newCard(s, domain.Receptive)
	yes := true
	a := choose(c, s.Items[1].ID, now.Add(-time.Minute)) // errou de verdade
	a.ClientCorrect, a.HasClientCorrect = &yes, true     // mas o app diz que acertou
	res := one(t, submitter(s, now), a)
	if res.Correct || s.Log[0].Correct || s.Log[0].ClientCorrect == nil || !*s.Log[0].ClientCorrect {
		t.Errorf("vale a correcao do servidor; a do cliente so fica registrada: %+v %+v", res, s.Log[0])
	}
}

func TestSubmitDuplicateDoesNothing(t *testing.T) {
	s := store("house")
	c := newCard(s, domain.Receptive)
	u := submitter(s, now)
	a := choose(c, c.ItemID, now.Add(-time.Minute))
	first := one(t, u, a)
	state := s.Cards[c.ID].State
	second := one(t, u, a)
	if first.Status != domain.Applied || second.Status != domain.Duplicate || len(s.Log) != 1 || s.Cards[c.ID].State != state {
		t.Fatalf("reenvio nao muda nada: %+v %+v %d", first, second, len(s.Log))
	}
	// o mesmo lote enviado de novo inteiro
	res, _ := u.Run(context.Background(), []domain.Answer{a, a})
	if res[0].Status != domain.Duplicate || res[1].Status != domain.Duplicate || s.Cards[c.ID].State != state {
		t.Errorf("lote repetido: %+v", res)
	}
}

func TestSubmitRejectsWhatCannotBeApplied(t *testing.T) {
	s := store("house", "cat")
	c := newCard(s, domain.Receptive)
	u := submitter(s, now)
	ok := choose(c, c.ItemID, now.Add(-time.Minute))
	cases := map[string]struct {
		mutate func(a *domain.Answer)
		reason domain.Reason
	}{
		"cartao inexistente":             {func(a *domain.Answer) { a.CardID = uuid.New() }, domain.ReasonCardNotFound},
		"exercicio errado para o cartao": {func(a *domain.Answer) { a.Exercise = domain.Type; a.Text = "house" }, domain.ReasonInvalidAnswer},
		"escolher sem item":              {func(a *domain.Answer) { a.ChosenItemID = uuid.Nil }, domain.ReasonInvalidAnswer},
		"respondeu antes de ver":         {func(a *domain.Answer) { a.ShownAt = a.AnsweredAt.Add(time.Second) }, domain.ReasonInvalidTime},
		"resposta do futuro":             {func(a *domain.Answer) { a.AnsweredAt = now.Add(time.Hour); a.ShownAt = now }, domain.ReasonInvalidTime},
	}
	for name, c := range cases {
		a := ok
		a.ID = uuid.New()
		c.mutate(&a)
		if res := one(t, u, a); res.Status != domain.Rejected || res.Reason != c.reason {
			t.Errorf("%s: %+v", name, res)
		}
	}
	bad := choose(c, c.ItemID, now.Add(-time.Minute))
	bad.Difficulty = "otimo"
	if res := one(t, u, bad); res.Reason != domain.ReasonInvalidAnswer {
		t.Errorf("dificuldade invalida: %+v", res)
	}
	if len(s.Log) != 0 || s.Cards[c.ID].State.State != srs.StateNew {
		t.Error("rejeitada nao grava nada nem muda o cartao")
	}
}

// Em branco e "nao sei": vale como erro e fica no log, senao o cartao nunca perderia forca.
func TestSubmitBlankTypedAnswerCountsAsAWrongOne(t *testing.T) {
	s := store("house")
	prod := newCard(s, domain.Productive)
	res := one(t, submitter(s, now), typed(prod, "   ", now.Add(-time.Minute)))
	if res.Status != domain.Applied || res.Correct || res.Grade != srs.Again || len(s.Log) != 1 {
		t.Errorf("digitar em branco: %+v", res)
	}
}

func TestSubmitRejectedAnswerDoesNotStopTheRest(t *testing.T) {
	s := store("house", "cat")
	c := newCard(s, domain.Receptive)
	bad := choose(c, c.ItemID, now.Add(-time.Minute))
	bad.CardID = uuid.New()
	good := choose(c, c.ItemID, now.Add(-time.Minute))
	res, err := submitter(s, now).Run(context.Background(), []domain.Answer{bad, good})
	if err != nil || res[0].Status != domain.Rejected || res[1].Status != domain.Applied {
		t.Fatalf("lote misto: %+v %v", res, err)
	}
}

func TestSubmitDefaultsAndLimits(t *testing.T) {
	s := store("house", "cat")
	// sem dificuldade informada: padrao medio
	c := newCard(s, domain.Receptive)
	a := choose(c, c.ItemID, now.Add(-time.Minute))
	a.Difficulty = ""
	if res := one(t, submitter(s, now), a); res.Grade != srs.Good {
		t.Errorf("padrao medio: %+v", res)
	}
	// dica limita a Hard; facil em reconhecimento vira Good
	c2 := newCard(s, domain.Receptive)
	hint := choose(c2, c2.ItemID, now.Add(-time.Minute))
	hint.Difficulty, hint.HintUsed = srs.DifficultyEasy, true
	if res := one(t, submitter(s, now), hint); res.Grade != srs.Hard {
		t.Errorf("dica limita a Hard: %+v", res)
	}
	c3 := newCard(s, domain.Receptive)
	easy := choose(c3, c3.ItemID, now.Add(-time.Minute))
	easy.Difficulty = srs.DifficultyEasy
	if res := one(t, submitter(s, now), easy); res.Grade != srs.Good {
		t.Errorf("facil em reconhecimento conta como Good: %+v", res)
	}
	// erro sem tipo: nao sabia
	c4 := newCard(s, domain.Receptive)
	miss := choose(c4, s.Items[1].ID, now.Add(-time.Minute))
	one(t, submitter(s, now), miss)
	if s.Log[len(s.Log)-1].ErrorKind != srs.ErrorNoIdea {
		t.Errorf("erro sem tipo vira no_idea: %+v", s.Log[len(s.Log)-1].ErrorKind)
	}
}

func TestSubmitRequiresASessionAndAValidBatch(t *testing.T) {
	s := store("house")
	c := newCard(s, domain.Receptive)
	anon := app.SubmitReviews{Viewers: &learningtest.Viewers{}, Repo: s, Scheduler: srs.SimpleV1{}}
	if _, err := anon.Run(context.Background(), []domain.Answer{choose(c, c.ItemID, now)}); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("sem sessao: %v", err)
	}
	u := submitter(s, now)
	if _, err := u.Run(context.Background(), nil); !errors.Is(err, app.ErrInvalidBatch) {
		t.Errorf("lote vazio: %v", err)
	}
	big := make([]domain.Answer, app.MaxBatch+1)
	if _, err := u.Run(context.Background(), big); !errors.Is(err, app.ErrInvalidBatch) {
		t.Errorf("lote grande demais: %v", err)
	}
	if _, err := u.Run(context.Background(), make([]domain.Answer, app.MaxBatch)); errors.Is(err, app.ErrInvalidBatch) {
		t.Error("200 respostas e o maximo permitido")
	}
}

func TestSubmitPropagatesRepositoryFailuresAndRollsBack(t *testing.T) {
	for _, method := range []string{"LockCard", "ExpectedTerm", "Settings", "InsertReview", "SaveCard", "CardLog"} {
		s := store("house", "cat")
		c := newCard(s, domain.Productive)
		// um cartao ja respondido, para a resposta antiga forcar o CardLog (replay)
		c.State = srs.CardState{ID: c.ID.String(), State: srs.StateReview, Stability: 2, Reps: 1, LastReviewAt: now, DueAt: now.Add(48 * time.Hour)}
		s.FailOn = method
		before := *s.Cards[c.ID]
		_, err := submitter(s, now.Add(time.Hour)).Run(context.Background(), []domain.Answer{typed(c, "house", now.Add(-time.Hour))})
		if !errors.Is(err, learningtest.ErrBoom) {
			t.Errorf("falha em %s deveria subir: %v", method, err)
		}
		if *s.Cards[c.ID] != before || len(s.Log) != 0 {
			t.Errorf("falha em %s: nada fica gravado", method)
		}
	}
}

func TestSubmitOutOfOrderAnswerReplaysTheWholeLog(t *testing.T) {
	s := store("house", "cat")
	c := newCard(s, domain.Receptive)
	u := submitter(s, now.Add(10*24*time.Hour))
	day := func(n int) time.Time { return now.Add(time.Duration(n) * 24 * time.Hour) }
	first := choose(c, c.ItemID, day(0))
	third := choose(c, c.ItemID, day(6))
	second := choose(c, c.ItemID, day(3)) // chega por ultimo, de outro aparelho
	for _, a := range []domain.Answer{first, third, second} {
		if res := one(t, u, a); res.Status != domain.Applied {
			t.Fatalf("aplicada: %+v", res)
		}
	}
	got := s.Cards[c.ID].State
	want, err := srs.Replay(srs.SimpleV1{NoFuzz: true}, c.ID.String(), []srs.Answer{
		{Review: srs.Review{Correct: true, Difficulty: srs.DifficultyMedium}, At: first.AnsweredAt},
		{Review: srs.Review{Correct: true, Difficulty: srs.DifficultyMedium}, At: second.AnsweredAt},
		{Review: srs.Review{Correct: true, Difficulty: srs.DifficultyMedium}, At: third.AnsweredAt},
	})
	if err != nil || got != want || len(s.Log) != 3 {
		t.Fatalf("o estado deve ser o do log em ordem de answered_at:\n got  %+v\n want %+v (%v) log=%d", got, want, err, len(s.Log))
	}
}
