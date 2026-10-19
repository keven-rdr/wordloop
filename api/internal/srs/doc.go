// Package srs e a biblioteca pura de repeticao espacada (docs/fase-1/01-srs.md): sem I/O, sem banco
// e sem relogio proprio. O agendador recebe o estado do cartao, a resposta e o instante e devolve o
// novo estado; trocar de estrategia e reproduzir o log de revisoes (Replay) na nova implementacao.
package srs
