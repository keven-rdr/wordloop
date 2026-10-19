// Command srs-sim roda o simulador de alunos virtuais e imprime o relatorio comparativo das estrategias
// (retencao real x alvo, carga de revisoes, palavras sabidas nos dias 30 e 90 e calibracao do R previsto).
// Uso: go run ./cmd/srs-sim [-days 90] [-new 8] [-budget 60] [-seed 1]
package main

import (
	"flag"
	"os"

	"github.com/keven-rdr/wordloop/api/internal/srs"
	"github.com/keven-rdr/wordloop/api/internal/srs/sim"
)

func main() {
	cfg := sim.DefaultConfig()
	flag.IntVar(&cfg.Days, "days", cfg.Days, "dias simulados")
	flag.IntVar(&cfg.NewPerDay, "new", cfg.NewPerDay, "palavras novas por dia de estudo")
	flag.IntVar(&cfg.Budget, "budget", cfg.Budget, "revisoes maximas por dia de estudo")
	flag.Uint64Var(&cfg.Seed, "seed", cfg.Seed, "semente do acaso")
	flag.Parse()
	sim.Report(os.Stdout, sim.Compare(cfg, []srs.Scheduler{srs.SimpleV1{}, sim.Ladder{}}))
}
