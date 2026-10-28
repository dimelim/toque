package update

import (
	"encoding/json"
	"testing"

	"toque/internal/i18n"
)

func TestNotesInFallsBackToSpanish(t *testing.T) {
	var m Manifest
	raw := `{"version":"1.19.0","notes":"Notas.","notesPt":"Notas em português.","notesEn":""}`
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	if got := m.NotesIn(i18n.PT); got != "Notas em português." {
		t.Fatalf("pt notes = %q", got)
	}
	if got := m.NotesIn(i18n.EN); got != "Notas." {
		t.Fatalf("missing en notes should fall back to Spanish, got %q", got)
	}
	if got := m.NotesIn(i18n.ES); got != "Notas." {
		t.Fatalf("es notes = %q", got)
	}
}
