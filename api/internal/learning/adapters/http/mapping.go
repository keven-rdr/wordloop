package httpapi

import (
	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

const mediaPath = "/media/"

func value[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

func planRequest(b gen.PlanRequest) app.PlanRequest {
	r := app.PlanRequest{Minutes: int(b.Minutes), TZ: value(b.Tz)}
	if b.Mode != nil {
		r.Mode = domain.Mode(*b.Mode)
	}
	if b.Locale != nil {
		r.Locale = string(*b.Locale)
	}
	return r
}

func studyPack(p domain.Pack) gen.StudyPack {
	cards := make([]gen.StudyCard, 0, len(p.Cards))
	for _, c := range p.Cards {
		cards = append(cards, studyCard(c))
	}
	return gen.StudyPack{SessionId: p.SessionID, Cards: cards, NewLimit: p.NewLimit, ReviewLimit: p.ReviewLimit, ExpiresAt: p.ExpiresAt}
}

func studyCard(c domain.PackCard) gen.StudyCard {
	out := gen.StudyCard{
		Id: c.Card.ID, ItemId: c.Card.ItemID, Direction: gen.StudyCardDirection(c.Card.Direction),
		State: gen.StudyCardState(c.Card.State.State), Exercise: gen.StudyCardExercise(c.Card.Exercise()),
		Term: c.Content.Term, Gloss: c.Content.Gloss, Mastery: float32(c.Mastery),
	}
	if c.Content.Example != "" {
		out.Example = &gen.ItemExample{Text: c.Content.Example, Translation: optional(c.Content.ExampleGloss)}
	}
	if c.Content.ImageKey != "" {
		url := mediaPath + c.Content.ImageKey
		out.ImageUrl = &url
	}
	if len(c.Options) > 0 {
		opts := make([]gen.ChoiceOption, 0, len(c.Options))
		for _, o := range c.Options {
			opts = append(opts, gen.ChoiceOption{ItemId: o.ItemID, Gloss: o.Gloss})
		}
		out.Options = &opts
	}
	return out
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func answers(in []gen.ReviewInput) []domain.Answer {
	out := make([]domain.Answer, 0, len(in))
	for _, r := range in {
		out = append(out, answer(r))
	}
	return out
}

func answer(r gen.ReviewInput) domain.Answer {
	a := domain.Answer{
		ID: r.Id, CardID: r.CardId, SessionID: value(r.SessionId), DeviceID: value(r.DeviceId),
		Exercise: domain.Exercise(r.Exercise), Source: string(value(r.Source)), ShownAt: r.ShownAt, AnsweredAt: r.AnsweredAt,
		ResponseMS: value(r.ResponseMs), Text: value(r.Answer), ChosenItemID: value(r.ChosenItemId),
		HintUsed: value(r.HintUsed), Explained: value(r.ExplanationViewed), ImageBefore: value(r.ImageBeforeAnswer),
		DifficultyHint: string(value(r.DifficultySuggested)), Difficulty: srs.Difficulty(value(r.Difficulty)),
		ErrorKind: srs.ErrorKind(value(r.ErrorKind)), ClientSeq: int64(value(r.ClientSeq)),
	}
	if r.ClientCorrect != nil {
		a.ClientCorrect, a.HasClientCorrect = r.ClientCorrect, true
	}
	return a
}

func batchResult(rs []domain.Result) gen.ReviewBatchResult {
	out := make([]gen.ReviewResult, 0, len(rs))
	for _, r := range rs {
		out = append(out, reviewResult(r))
	}
	return gen.ReviewBatchResult{Results: out}
}

func reviewResult(r domain.Result) gen.ReviewResult {
	out := gen.ReviewResult{Id: r.ID, Status: gen.ReviewResultStatus(r.Status)}
	if r.Status == domain.Rejected {
		reason := gen.ReviewResultReason(r.Reason)
		out.Reason = &reason
	}
	if r.Status == domain.Applied {
		grade, correct, typo := int(r.Grade), r.Correct, r.Typo
		out.Grade, out.Correct, out.Typo = &grade, &correct, &typo
	}
	if r.Card != nil {
		p := progress(*r.Card)
		out.Card = &p
	}
	return out
}

func progress(c domain.Card) gen.CardProgress {
	leech := c.State.Leech()
	return gen.CardProgress{
		Id: c.ID, State: gen.CardProgressState(c.State.State), Stability: float32(c.State.Stability),
		DueAt: c.State.DueAt, Lapses: c.State.Lapses, Mastery: float32(srs.Mastery(c.State.Stability)), Leech: &leech,
	}
}
