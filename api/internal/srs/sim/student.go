// Package sim e o simulador de alunos virtuais do SRS (docs/fase-1/01-srs.md, secao 7): cada aluno esquece
// segundo uma curva PROPRIA (verdade), diferente da curva do modelo, para nao avaliar o algoritmo por ele mesmo.
// So para testes e relatorios; nao roda em producao.
package sim

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// Profile descreve um aluno virtual.
type Profile struct {
	Name      string
	Base      float64 // estabilidade verdadeira inicial media de uma palavra (dias)
	Gain      float64 // multiplicador de estabilidade a cada recordacao bem-sucedida
	Diligence float64 // probabilidade de estudar num dia
	Noise     float64 // ruido da autoavaliacao de dificuldade (desvio, de 0 a 1)
}

const (
	failShrink    = 0.35 // ao esquecer, a estabilidade verdadeira cai para 35%
	minTrueS      = 0.5
	firstRecall   = 0.85 // chance de acertar o teste logo depois da apresentacao (reconhecimento)
	hardBelow     = 0.75 // autoavaliacao: abaixo disso "dificil"
	mediumBelow   = 0.93 // abaixo disso "medio"; senao "facil"
	trueRetention = 0.9  // a verdade tambem passa por 90% quando t = S verdadeiro, mas com forma exponencial
)

// memory e a memoria VERDADEIRA de um item: p(t) = 0,9^(t/S), uma exponencial (o modelo usa potencia).
type memory struct {
	s    float64
	last time.Time
}

func (m memory) recall(now time.Time) float64 {
	days := math.Max(0, now.Sub(m.last).Hours()/24)
	return math.Pow(trueRetention, days/m.s)
}

// reinforce atualiza a memoria depois de uma tentativa: o efeito de espacamento faz lembrar com p baixo
// reforcar mais; esquecer encolhe a estabilidade.
func (m *memory) reinforce(p float64, ok bool, gain float64, now time.Time) {
	if ok {
		m.s *= math.Max(1.1, 1+(gain-1)*(1.5-p))
	} else {
		m.s = math.Max(minTrueS, m.s*failShrink)
	}
	m.last = now
}

// student junta o perfil, a memoria verdadeira de cada item e o gerador de acaso (semente fixa).
type student struct {
	p   Profile
	rng *rand.Rand
	mem map[string]*memory
}

func newStudent(p Profile, seed uint64) *student {
	// Acaso de simulacao com semente fixa (resultado reproduzivel): nao e seguranca.
	rng := rand.New(rand.NewPCG(seed, 0x5eed)) //nolint:gosec // G404: simulacao determinista
	return &student{p: p, rng: rng, mem: map[string]*memory{}}
}

// learn apresenta uma palavra nova: cada uma tem a sua dificuldade (0,6x a 1,6x a base).
func (s *student) learn(id string, now time.Time) {
	s.mem[id] = &memory{s: s.p.Base * (0.6 + s.rng.Float64()), last: now}
}

// studies diz se o aluno estuda hoje.
func (s *student) studies() bool { return s.rng.Float64() < s.p.Diligence }

// answer devolve se acertou e a resposta que o app recebe (com a dificuldade que o aluno acha que sentiu).
func (s *student) answer(id string, c srs.CardState, now time.Time) (srs.Review, float64) {
	m := s.mem[id]
	p := firstRecall
	if c.State != srs.StateNew {
		p = m.recall(now)
	}
	ok := s.rng.Float64() < p
	m.reinforce(p, ok, s.p.Gain, now)
	return srs.Review{
		Correct: ok, Difficulty: s.perceive(p), ErrorKind: srs.ErrorNoIdea,
		Production: c.Stability >= 7, // digitar a partir de S >= 7 d (docs/fase-1, secao 9)
	}, p
}

func (s *student) perceive(p float64) srs.Difficulty {
	x := p + s.rng.NormFloat64()*s.p.Noise
	switch {
	case x < hardBelow:
		return srs.DifficultyHard
	case x < mediumBelow:
		return srs.DifficultyMedium
	}
	return srs.DifficultyEasy
}

// known conta quantas palavras o aluno lembraria agora com pelo menos 90% de chance.
func (s *student) known(ids []string, now time.Time) int {
	n := 0
	for _, id := range ids {
		if s.mem[id].recall(now) >= trueRetention {
			n++
		}
	}
	return n
}
