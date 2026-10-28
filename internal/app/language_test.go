package app

import (
	"errors"
	"os"
	"strings"
	"testing"

	"toque/internal/i18n"
)

func TestDetectLanguage(t *testing.T) {
	country := func(c string, err error) func() (string, error) {
		return func() (string, error) { return c, err }
	}
	offline := errors.New("offline")
	cases := []struct {
		tag     string
		country func() (string, error)
		want    i18n.Lang
		sure    bool
	}{
		{"es-MX", country("BR", nil), i18n.ES, true},
		{"pt-BR", country("US", nil), i18n.PT, true},
		{"en-US", country("BR", nil), i18n.PT, true},
		{"en-GB", country("AR", nil), i18n.ES, true},
		{"en-US", country("US", nil), i18n.EN, true},
		{"fr-FR", country("FR", nil), i18n.EN, true},
		{"de-DE", country("MX", nil), i18n.ES, true},
		{"en-US", country("", offline), i18n.EN, false},
		{"", country("", offline), i18n.EN, false},
	}
	for _, c := range cases {
		got, sure := detectLanguage(c.tag, c.country)
		if got != c.want || sure != c.sure {
			t.Errorf("detectLanguage(%q) = %s, %v, want %s, %v", c.tag, got, sure, c.want, c.sure)
		}
	}
}

func TestLanguageChoiceWinsOverDetection(t *testing.T) {
	f := languageFile{Detected: "pt"}
	if f.active() != i18n.PT || f.info().Choice != autoLanguage {
		t.Fatalf("auto should follow detection, got %+v", f.info())
	}
	f.Choice = "en"
	if f.active() != i18n.EN || f.info().Detected != "pt" {
		t.Fatalf("a manual choice should win, got %+v", f.info())
	}
}

func TestGuessedLanguageIsNotSaved(t *testing.T) {
	orig := systemDir
	defer func() { systemDir = orig }()
	dir := t.TempDir()
	systemDir = func() string { return dir }
	if err := saveLanguage(languageFile{Choice: "pt", Detected: "en", guess: true}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(languagePath())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "detected") || !strings.Contains(string(b), `"choice":"pt"`) {
		t.Fatalf("only the manual choice should be saved when detection was a guess, got %s", b)
	}
}
