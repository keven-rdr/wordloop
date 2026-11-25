package domain

import "testing"

func TestReviewPointsByExerciseAndLevel(t *testing.T) {
	cases := []struct {
		name       string
		correct    bool
		production bool
		cefr       string
		want       int
	}{
		{"acerto no reconhecimento A1", true, false, "A1", 10},
		{"acerto na producao A1", true, true, "A1", 15},
		{"sem nivel conta como A1", true, false, "", 10},
		{"A2", true, false, "A2", 12},
		{"B1 na producao arredonda (22,5)", true, true, "B1", 23},
		{"B2 ou mais dobra", true, false, "C1", 20},
		{"erro vale o esforco, sem nivel", false, true, "C1", 2},
	}
	for _, c := range cases {
		if got := ReviewPoints(c.correct, c.production, c.cefr); got != c.want {
			t.Errorf("%s: %d, esperava %d", c.name, got, c.want)
		}
	}
}

func TestDayDeltaCarriesThePointsOfTheAnswer(t *testing.T) {
	r := LogRow{Correct: true, Production: true, Level: "A2"}
	if d := r.DayDelta(); d.Points != 18 {
		t.Errorf("15 x 1,2 = 18: %+v", d)
	}
}
