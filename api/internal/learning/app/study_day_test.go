package app_test

import (
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

func TestSubmitSumsEachAnswerIntoTheStudyDayOnce(t *testing.T) {
	s := store("house", "cat")
	c := newCard(s, domain.Receptive)
	u := submitter(s, now)
	a := choose(c, c.ItemID, now.Add(-time.Minute))
	one(t, u, a)
	one(t, u, a) // reenvio do mesmo UUID: nao soma de novo
	if len(s.Days) != 1 {
		t.Fatalf("uma resposta aplicada soma uma vez: %d", len(s.Days))
	}
	if got := s.Days[0].DayDelta(); got.Correct != 1 || got.NewItems != 1 || got.Lapses != 0 {
		t.Errorf("1o acerto de um cartao novo: %+v", got)
	}
}

func TestSubmitCountsALapseWhenAKnownCardIsMissed(t *testing.T) {
	s := store("house", "cat")
	c := newCard(s, domain.Receptive)
	c.State = srs.CardState{ID: c.ID.String(), State: srs.StateReview, Stability: 10, Reps: 3, LastReviewAt: now.Add(-72 * time.Hour), DueAt: now.Add(-time.Hour)}
	one(t, submitter(s, now), choose(c, s.Items[1].ID, now.Add(-time.Minute)))
	got := s.Days[0].DayDelta()
	if got.Correct != 0 || got.Lapses != 1 || got.NewItems != 0 {
		t.Errorf("erro em cartao conhecido e lapso, nao palavra nova: %+v", got)
	}
}

func TestRejectedAnswersDoNotTouchTheStudyDay(t *testing.T) {
	s := store("house")
	c := newCard(s, domain.Receptive)
	a := choose(c, c.ItemID, now.Add(time.Hour)) // resposta do futuro
	one(t, submitter(s, now), a)
	if len(s.Days) != 0 {
		t.Errorf("rejeitada nao entra no dia: %d", len(s.Days))
	}
}
