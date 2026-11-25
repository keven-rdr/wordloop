//go:build integration

package postgres_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

func TestExportBringsEverythingOfThePersonAndNoSecret(t *testing.T) {
	f := setup(t)
	ana := f.person(t, "Ana", "ana@example.com")
	f.person(t, "Bia", "bia@example.com") // os dados de outra pessoa nao entram
	if _, err := f.pool.Exec(context.Background(), `INSERT INTO leaderboard_profile (user_id, nickname, visible) VALUES ($1, 'ana_1', true)`, ana); err != nil {
		t.Fatal(err)
	}

	out, err := f.store.Export(context.Background(), ana)
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(out.Data, &data); err != nil {
		t.Fatalf("o banco deve devolver um objeto JSON: %v", err)
	}
	for _, key := range []string{"profile", "logins", "passkeys", "sessions", "preferences", "goals", "study_days", "study_sessions", "cards", "reviews"} {
		if _, ok := data[key]; !ok {
			t.Errorf("falta a secao %q", key)
		}
	}
	sizes := map[string]int{"logins": 1, "passkeys": 1, "sessions": 1, "goals": 1, "study_days": 1, "study_sessions": 1, "cards": 1, "reviews": 2}
	for key, want := range sizes {
		var list []json.RawMessage
		if err := json.Unmarshal(data[key], &list); err != nil || len(list) != want {
			t.Errorf("%s: %d itens, esperava %d (%v)", key, len(list), want, err)
		}
	}

	var profile map[string]any
	if err := json.Unmarshal(data["profile"], &profile); err != nil || profile["email"] != "ana@example.com" || profile["display_name"] != "Ana" || profile["has_password"] != true {
		t.Errorf("perfil: %v %v", profile, err)
	}
	var prefs map[string]any
	if err := json.Unmarshal(data["preferences"], &prefs); err != nil || prefs["theme"] != "dark" {
		t.Errorf("preferencias: %v %v", prefs, err)
	}
	var ranking map[string]any
	if err := json.Unmarshal(data["ranking"], &ranking); err != nil || ranking["nickname"] != "ana_1" || ranking["visible"] != true {
		t.Errorf("perfil do ranking: %v %v", ranking, err)
	}
	var cards []map[string]any
	if err := json.Unmarshal(data["cards"], &cards); err != nil || cards[0]["term"] != "house" || cards[0]["item"] != "word|house|noun|1" {
		t.Errorf("cartao com a palavra: %v %v", cards, err)
	}
	if strings.Contains(string(out.Data), "bia@example.com") || strings.Contains(string(out.Data), "sub-Bia") {
		t.Error("a exportacao de uma pessoa nao pode trazer dado de outra")
	}
}

// O coracao da LGPD aqui: nada de segredo sai no arquivo, nem por nome de coluna nem por valor.
func TestExportNeverLeaksASecret(t *testing.T) {
	f := setup(t)
	ana := f.person(t, "Ana", "ana@example.com")
	out, err := f.store.Export(context.Background(), ana)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out.Data)
	for _, key := range []string{"password_hash", "public_key", "credential_id", "id_hash", "refresh_token_enc", "failed_logins", "locked_until"} {
		if strings.Contains(text, `"`+key+`"`) {
			t.Errorf("coluna sensivel %q na exportacao", key)
		}
	}
	for _, secret := range []string{canaryPasswordHash, canaryRefreshToken, canaryPublicKey, canarySessionHash, canaryTokenHash} {
		if strings.Contains(text, secret) || strings.Contains(text, hex.EncodeToString([]byte(secret))) {
			t.Errorf("valor secreto %q na exportacao", secret)
		}
	}
}

func TestExportOfAPersonWithoutStudyIsStillACompleteObject(t *testing.T) {
	f := setup(t)
	var id string
	if err := f.pool.QueryRow(context.Background(), `INSERT INTO app_user (display_name, email) VALUES ('Nova', 'nova@example.com') RETURNING id::text`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	var user = mustUUID(t, id)
	out, err := f.store.Export(context.Background(), user)
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(out.Data, &data); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"logins", "passkeys", "sessions", "goals", "study_days", "study_sessions", "cards", "reviews"} {
		if string(data[key]) != "[]" {
			t.Errorf("%s sem dados deve ser lista vazia, nao %s", key, data[key])
		}
	}
}
