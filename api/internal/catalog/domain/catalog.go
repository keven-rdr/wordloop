// Package domain guarda os tipos do catalogo: o que existe para estudar e quem pode ver o que.
package domain

import "github.com/google/uuid"

// Viewer e quem pergunta: so a sessao decide, nunca o pedido.
type Viewer struct {
	Authenticated bool
	Admin         bool
}

// SeesDrafts diz se o rascunho aparece: so para administradores.
func (v Viewer) SeesDrafts() bool { return v.Admin }

type Sort string

const (
	TermAsc  Sort = "term,asc"
	TermDesc Sort = "term,desc"
	CEFRAsc  Sort = "cefr,asc"
	CEFRDesc Sort = "cefr,desc"
)

func (s Sort) Valid() bool {
	switch s {
	case TermAsc, TermDesc, CEFRAsc, CEFRDesc:
		return true
	}
	return false
}

// Query descreve uma listagem. Campos vazios nao filtram.
type Query struct {
	Page, Size int
	Sort       Sort
	Text       string
	Unit       uuid.UUID
	CEFR, Kind string
	Locale     string
	Drafts     bool
	ID         uuid.UUID
}

type Summary struct {
	ID       uuid.UUID
	Kind     string
	Term     string
	Gloss    string
	CEFR     string
	UnitID   uuid.UUID
	Status   string
	ImageKey string
}

type Page struct {
	Items             []Summary
	Page, Size, Total int
}

type Example struct{ Text, Translation string }

type Image struct{ Key, Credit string }

type Related struct {
	ID          uuid.UUID
	Term, Gloss string
	Type        string
}

type Detail struct {
	Summary
	Examples []Example
	Images   []Image
	Related  []Related
}

type Unit struct {
	ID       uuid.UUID
	Position int
	ThemeKey string
	Items    int
}

type Credit struct{ Code, Name, URL, License, Notice string }
