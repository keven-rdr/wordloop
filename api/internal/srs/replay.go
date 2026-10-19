package srs

import "time"

// Answer e uma linha do log de revisoes: a resposta corrigida e o instante em que foi dada.
type Answer struct {
	Review
	At time.Time
}

// Replay reproduz o log (em ordem de At) numa estrategia e devolve o estado final. E assim que se troca de
// estrategia e que se confere que o estado guardado no cartao e so um cache do log imutavel.
func Replay(s Scheduler, cardID string, log []Answer) (CardState, error) {
	c := CardState{ID: cardID, State: StateNew}
	for _, a := range log {
		next, err := s.Next(c, a.Review, a.At)
		if err != nil {
			return c, err
		}
		c = next
	}
	return c, nil
}
