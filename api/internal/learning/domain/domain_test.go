package domain

import (
	"errors"
	"testing"
	"time"
)

func TestCorrectTyped(t *testing.T) {
	cases := map[string]struct {
		expected, answer string
		want             Verdict
	}{
		"exato":                    {"house", "house", Verdict{Correct: true}},
		"maiusculas e espacos":     {"House", "  HOUSE ", Verdict{Correct: true}},
		"acentos":                  {"café", "cafe", Verdict{Correct: true}},
		"expressao com espacos":    {"give up", "give   up", Verdict{Correct: true}},
		"vazio":                    {"house", "   ", Verdict{}},
		"troca de uma letra":       {"house", "hause", Verdict{Typo: true}},
		"letra a mais":             {"house", "housse", Verdict{Typo: true}},
		"letra a menos":            {"house", "hous", Verdict{Typo: true}},
		"inversao de vizinhas":     {"house", "huose", Verdict{Typo: true}},
		"letra a menos no inicio":  {"house", "ouse", Verdict{Typo: true}},
		"duas diferencas":          {"house", "hxuxe", Verdict{}},
		"palavra curta nao tolera": {"cat", "cut", Verdict{}},
		"quatro letras nao tolera": {"door", "dorr", Verdict{}},
		"outra palavra":            {"house", "mouse", Verdict{Typo: true}},
		"muito diferente":          {"house", "elephant", Verdict{}},
		"tamanhos distantes":       {"house", "housess", Verdict{}},
	}
	for name, c := range cases {
		if got := CorrectTyped(c.expected, c.answer); got != c.want {
			t.Errorf("%s: CorrectTyped(%q, %q) = %+v, esperava %+v", name, c.expected, c.answer, got, c.want)
		}
	}
}

func TestPlanLimits(t *testing.T) {
	cases := []struct {
		minutes int
		mode    Mode
		due     int
		want    Limits
	}{
		{5, ModeMixed, 0, Limits{Reviews: 20, New: 1}},
		{15, ModeMixed, 10, Limits{Reviews: 60, New: 3}},
		{30, ModeMixed, 0, Limits{Reviews: 120, New: 6}},
		{60, ModeMixed, 0, Limits{Reviews: 240, New: 8}}, // 60/5 = 12, mas o teto inicial e 8
		{15, ModeReview, 0, Limits{Reviews: 60, New: 0}},
		{15, ModeLearn, 0, Limits{Reviews: 0, New: 3}},
		{15, ModeMixed, 121, Limits{Reviews: 60, New: 0}}, // fila acima de 2x o limite: sem novas
		{15, ModeMixed, 120, Limits{Reviews: 60, New: 3}}, // exatamente 2x ainda entra
	}
	for _, c := range cases {
		got, err := PlanLimits(c.minutes, c.mode, c.due)
		if err != nil || got != c.want {
			t.Errorf("PlanLimits(%d, %s, %d) = %+v %v, esperava %+v", c.minutes, c.mode, c.due, got, err, c.want)
		}
	}
	if _, err := PlanLimits(10, ModeMixed, 0); !errors.Is(err, ErrInvalidMinutes) {
		t.Errorf("10 minutos: %v", err)
	}
	if _, err := PlanLimits(15, "tudo", 0); !errors.Is(err, ErrInvalidMode) {
		t.Errorf("modo invalido: %v", err)
	}
}

func TestLogicalDateTurnsAtTheCutoffInTheUsersTimezone(t *testing.T) {
	sp := "America/Sao_Paulo" // UTC-3, sem horario de verao desde 2019
	cases := map[string]struct {
		at   time.Time
		want string
	}{
		"de tarde":                   {time.Date(2026, 10, 19, 17, 0, 0, 0, time.UTC), "2026-10-19"}, // 14:00 local
		"de madrugada ainda e ontem": {time.Date(2026, 10, 19, 5, 0, 0, 0, time.UTC), "2026-10-18"},  // 02:00 local
		"depois do corte e hoje":     {time.Date(2026, 10, 19, 7, 30, 0, 0, time.UTC), "2026-10-19"}, // 04:30 local
		"no corte exato":             {time.Date(2026, 10, 19, 7, 0, 0, 0, time.UTC), "2026-10-19"},  // 04:00 local
	}
	for name, c := range cases {
		got, err := LogicalDate(c.at, sp, 4)
		if err != nil || got.Format("2006-01-02") != c.want {
			t.Errorf("%s: %v %v, esperava %s", name, got, err, c.want)
		}
	}
	utc, _ := LogicalDate(time.Date(2026, 10, 19, 3, 0, 0, 0, time.UTC), "UTC", 4)
	if utc.Format("2006-01-02") != "2026-10-18" {
		t.Errorf("03:00 UTC com corte de 4 h ainda e o dia anterior: %v", utc)
	}
}

func TestInvalidTimezone(t *testing.T) {
	for _, tz := range []string{"", "Marte/Olympus", "Local"} {
		if ValidTZ(tz) {
			t.Errorf("%q nao e um fuso valido", tz)
		}
	}
	if _, err := LogicalDate(time.Now(), "Marte/Olympus", 4); !errors.Is(err, ErrInvalidTZ) {
		t.Errorf("LogicalDate com fuso invalido: %v", err)
	}
	if !ValidTZ("Europe/Lisbon") || !ValidTZ("America/Sao_Paulo") {
		t.Error("fusos IANA validos")
	}
}

func TestViewerAndCardBasics(t *testing.T) {
	if (Viewer{}).SeesDrafts() || !(Viewer{Admin: true}).SeesDrafts() {
		t.Error("so o admin ve rascunho")
	}
	if (Card{Direction: Receptive}).Exercise() != Choose || (Card{Direction: Productive}).Exercise() != Type {
		t.Error("receptivo escolhe, produtivo digita")
	}
	if !Expiry(time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)).After(time.Date(2026, 10, 19, 6, 0, 0, 0, time.UTC)) {
		t.Error("o pacote vale por varias horas")
	}
}
