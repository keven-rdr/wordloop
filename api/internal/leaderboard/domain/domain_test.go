package domain

import (
	"errors"
	"testing"
	"time"
)

func date(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestLevelCurve(t *testing.T) {
	cases := map[int]Level{
		0:    {1, 0, 100},
		99:   {1, 0, 100},
		100:  {2, 100, 250},
		249:  {2, 100, 250},
		250:  {3, 250, 450},
		450:  {4, 450, 700},
		5000: LevelOf(5000),
	}
	for total, want := range cases {
		if got := LevelOf(total); got != want || total < got.Floor || total >= got.Next {
			t.Errorf("LevelOf(%d) = %+v, esperava %+v", total, got, want)
		}
	}
}

func TestNicknameRules(t *testing.T) {
	for _, ok := range []string{"ana", "Ana_Silva", "keven2026", "abcdefghijklmnopqrst"} {
		if err := CheckNickname(ok); err != nil {
			t.Errorf("%q deveria valer: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "ab", "abcdefghijklmnopqrstu", "com espaco", "joão", "ana!", "SuperAdmin", "xxPUTAxx"} {
		if err := CheckNickname(bad); !errors.Is(err, ErrInvalidNickname) {
			t.Errorf("%q deveria ser recusado: %v", bad, err)
		}
	}
}

func TestPeriodRanges(t *testing.T) {
	from, to, err := Week.Range(date("2026-11-27")) // sexta
	if err != nil || from != date("2026-11-23") || to != date("2026-11-29") {
		t.Errorf("semana: %v %v %v", from, to, err)
	}
	from, to, err = All.Range(date("2026-11-27"))
	if err != nil || !from.Before(date("2000-01-01")) || !to.After(date("2100-01-01")) {
		t.Errorf("geral: %v %v %v", from, to, err)
	}
	if _, _, err := Period("month").Range(date("2026-11-27")); !errors.Is(err, ErrInvalidPeriod) {
		t.Errorf("periodo desconhecido: %v", err)
	}
}

func TestTodayFollowsTheTimezoneAndTheCutoff(t *testing.T) {
	at := time.Date(2026, 11, 23, 5, 30, 0, 0, time.UTC) // 02:30 em Sao Paulo: antes do corte das 04:00
	if got := Today(at, "America/Sao_Paulo", 4); got != date("2026-11-22") {
		t.Errorf("madrugada conta no dia anterior: %v", got)
	}
	if got := Today(at, "", 0); got != date("2026-11-23") {
		t.Errorf("sem fuso, UTC: %v", got)
	}
}
