package domain

import (
	"time"

	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// LogicalDate e o "dia de estudo" de um instante: o dia local descontado o corte (04:00 por padrao), para
// o estudo da madrugada nao quebrar a sequencia. Devolve a data (meia-noite UTC) para gravar como `date`.
func LogicalDate(t time.Time, tz string, cutoffHour int) (time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil || tz == "" {
		return time.Time{}, ErrInvalidTZ
	}
	local := t.In(loc).Add(-time.Duration(cutoffHour) * time.Hour)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC), nil
}

// ValidTZ diz se o fuso e um nome IANA conhecido.
func ValidTZ(tz string) bool {
	_, err := time.LoadLocation(tz)
	return err == nil && tz != "" && tz != "Local"
}

// DayDelta e o que uma resposta soma no dia de estudo (study_day).
type DayDelta struct {
	ActiveMS                  int64
	Correct, Lapses, NewItems int
}

// DayDelta resume a resposta para a agregacao diaria: tempo ativo, acerto, lapso (o cartao perdeu forca) e
// palavra nova (1a revisao do cartao).
func (r LogRow) DayDelta() DayDelta {
	d := DayDelta{ActiveMS: int64(r.ResponseMS)}
	if r.Correct {
		d.Correct = 1
	}
	if r.After.Lapses > r.Before.Lapses {
		d.Lapses = 1
	}
	if r.Before.State == srs.StateNew || r.Before.State == "" {
		d.NewItems = 1
	}
	return d
}
