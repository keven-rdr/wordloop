package app

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

const clockSkew = 10 * time.Minute // tolerancia do relogio do aparelho para respostas "do futuro"

// decided e uma resposta validada e corrigida, pronta para gravar.
type decided struct {
	card    domain.Card
	verdict domain.Verdict
	review  srs.Review
	grade   srs.Grade
}

// apply processa uma resposta dentro da transacao: valida, corrige, calcula o novo estado e grava.
func (u SubmitReviews) apply(ctx context.Context, tx Tx, v domain.Viewer, a domain.Answer) (domain.Result, error) {
	d, why, err := u.decide(ctx, tx, v, a)
	if err != nil {
		return domain.Result{}, err
	}
	if why != "" {
		return rejected(a.ID, why), nil
	}
	return u.commit(ctx, tx, v, a, d)
}

// decide valida a resposta e a corrige no servidor; um motivo nao vazio e uma rejeicao (nao um erro).
func (u SubmitReviews) decide(ctx context.Context, tx Tx, v domain.Viewer, a domain.Answer) (decided, domain.Reason, error) {
	if a.AnsweredAt.Before(a.ShownAt) || a.AnsweredAt.After(u.Now.now().Add(clockSkew)) {
		return decided{}, domain.ReasonInvalidTime, nil
	}
	card, ok, err := tx.LockCard(ctx, v.UserID, a.CardID)
	if err != nil || !ok {
		return decided{}, domain.ReasonCardNotFound, err
	}
	verdict, ok, err := u.correct(ctx, tx, card, a)
	if err != nil || !ok {
		return decided{}, domain.ReasonInvalidAnswer, err
	}
	rev := review(a, verdict)
	grade, err := rev.Grade()
	if err != nil {
		return decided{}, domain.ReasonInvalidAnswer, nil
	}
	return decided{card: card, verdict: verdict, review: rev, grade: grade}, "", nil
}

// commit calcula o estado novo, grava a linha do log (so se o UUID e inedito) e atualiza o cartao.
func (u SubmitReviews) commit(ctx context.Context, tx Tx, v domain.Viewer, a domain.Answer, d decided) (domain.Result, error) {
	after, err := u.nextState(ctx, tx, d.card, a, d.review)
	if err != nil {
		return domain.Result{}, err
	}
	settings, err := tx.Settings(ctx, v.UserID)
	if err != nil {
		return domain.Result{}, err
	}
	row, err := u.logRow(a, v, settings, d, after)
	if err != nil {
		return rejected(a.ID, domain.ReasonInvalidTime), nil
	}
	inserted, err := tx.InsertReview(ctx, row)
	if err != nil {
		return domain.Result{}, err
	}
	card := d.card
	if !inserted {
		return domain.Result{ID: a.ID, Status: domain.Duplicate, Card: &card}, nil
	}
	card.State, card.Scheduler = after, u.Scheduler.Name()
	if err := tx.SaveCard(ctx, card); err != nil {
		return domain.Result{}, err
	}
	return domain.Result{
		ID: a.ID, Status: domain.Applied, Correct: d.verdict.Correct, Typo: d.verdict.Typo, Grade: d.grade, Card: &card,
	}, nil
}

// correct refaz a correcao no servidor: digitar compara com o termo; escolher confere o item escolhido.
// O segundo retorno e false quando a resposta nao serve para o exercicio do cartao.
func (u SubmitReviews) correct(ctx context.Context, tx Tx, card domain.Card, a domain.Answer) (domain.Verdict, bool, error) {
	if a.Exercise != card.Exercise() {
		return domain.Verdict{}, false, nil
	}
	if a.Exercise == domain.Choose {
		return domain.Verdict{Correct: a.ChosenItemID == card.ItemID}, a.ChosenItemID != uuid.Nil, nil
	}
	term, err := tx.ExpectedTerm(ctx, card.ItemID)
	if err != nil {
		return domain.Verdict{}, false, err
	}
	// Resposta em branco e "nao sei": conta como erro (nao pode sumir do log, senao o cartao nunca perde forca).
	return domain.CorrectTyped(term, a.Text), true, nil
}

// review traduz a resposta ja corrigida no que o agendador entende.
func review(a domain.Answer, v domain.Verdict) srs.Review {
	r := srs.Review{
		Correct: v.Correct, Typo: v.Typo, Production: a.Exercise == domain.Type,
		Hint: a.HintUsed, Explained: a.Explained, ImageBefore: a.ImageBefore,
	}
	if v.Correct {
		r.Difficulty = a.Difficulty
		if r.Difficulty == "" {
			r.Difficulty = srs.DifficultyMedium
		}
	} else {
		r.ErrorKind = a.ErrorKind
		if r.ErrorKind == srs.ErrorNone {
			r.ErrorKind = srs.ErrorNoIdea
		}
	}
	return r
}

// nextState calcula o estado depois da resposta. Se ela chegou fora de ordem (anterior a ultima revisao do
// cartao, por exemplo de outro aparelho), reproduz o log inteiro com ela no lugar certo.
func (u SubmitReviews) nextState(ctx context.Context, tx Tx, card domain.Card, a domain.Answer, r srs.Review) (srs.CardState, error) {
	if !a.AnsweredAt.Before(card.State.LastReviewAt) {
		return u.Scheduler.Next(card.State, r, a.AnsweredAt)
	}
	log, err := tx.CardLog(ctx, card.ID)
	if err != nil {
		return srs.CardState{}, err
	}
	log = append(log, srs.Answer{Review: r, At: a.AnsweredAt})
	sort.SliceStable(log, func(i, j int) bool { return log[i].At.Before(log[j].At) })
	return srs.Replay(u.Scheduler, card.State.ID, log)
}

func (u SubmitReviews) logRow(a domain.Answer, v domain.Viewer, s domain.Settings, d decided, after srs.CardState) (domain.LogRow, error) {
	date, err := domain.LogicalDate(a.AnsweredAt, s.TZ, s.CutoffHour)
	if err != nil {
		return domain.LogRow{}, err
	}
	a.Difficulty, a.ErrorKind = d.review.Difficulty, d.review.ErrorKind // o que de fato valeu (com os padroes), para o replay
	before := d.card.State
	elapsed := 0.0
	if !before.LastReviewAt.IsZero() {
		elapsed = a.AnsweredAt.Sub(before.LastReviewAt).Hours() / 24
	}
	return domain.LogRow{
		Answer: a, UserID: v.UserID, TZ: s.TZ, CutoffHour: s.CutoffHour, LocalDate: date,
		Correct: d.verdict.Correct, Typo: d.verdict.Typo, Grade: d.grade, Before: before, After: after,
		RAtReview: u.Scheduler.Retrievability(before, a.AnsweredAt), ElapsedDays: elapsed,
		ScheduledDays: after.DueAt.Sub(a.AnsweredAt).Hours() / 24,
		Scheduler:     u.Scheduler.Name(), SchedulerVersion: SchedulerVersion,
		Production: a.Exercise == domain.Type,
	}, nil
}
