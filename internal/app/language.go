package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"

	"toque/internal/geo"
	"toque/internal/i18n"
)

const autoLanguage = "auto"

type LanguageInfo struct {
	Choice   string `json:"choice"`
	Active   string `json:"active"`
	Detected string `json:"detected"`
}

type languageFile struct {
	Choice   string `json:"choice,omitempty"`
	Detected string `json:"detected,omitempty"`
	guess    bool
}

var (
	portugueseCountries = map[string]bool{"BR": true, "PT": true, "AO": true, "MZ": true, "CV": true, "GW": true, "ST": true, "TL": true}
	spanishCountries    = map[string]bool{
		"ES": true, "MX": true, "AR": true, "CO": true, "CL": true, "PE": true, "VE": true, "EC": true, "GT": true, "CU": true, "BO": true,
		"DO": true, "HN": true, "PY": true, "SV": true, "NI": true, "CR": true, "PA": true, "UY": true, "PR": true, "GQ": true,
	}
)

func languagePath() string { return filepath.Join(systemDir(), "language.json") }

func windowsLanguage() string {
	langs, err := windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
	if err != nil || len(langs) == 0 {
		return ""
	}
	return langs[0]
}

func tagLanguage(tag string) (i18n.Lang, bool) {
	prefix, _, _ := strings.Cut(strings.ToLower(tag), "-")
	return i18n.Parse(prefix)
}

func countryLanguage(country string) (i18n.Lang, bool) {
	switch c := strings.ToUpper(country); {
	case portugueseCountries[c]:
		return i18n.PT, true
	case spanishCountries[c]:
		return i18n.ES, true
	}
	return "", false
}

func detectLanguage(windowsTag string, country func() (string, error)) (i18n.Lang, bool) {
	if l, ok := tagLanguage(windowsTag); ok && l != i18n.EN {
		return l, true
	}
	c, err := country()
	if l, ok := countryLanguage(c); ok && err == nil {
		return l, true
	}
	return i18n.EN, err == nil
}

func selfCountry() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	p, err := geo.Self(ctx)
	return p.Country, err
}

func (f languageFile) active() i18n.Lang {
	if l, ok := i18n.Parse(f.Choice); ok {
		return l
	}
	if l, ok := i18n.Parse(f.Detected); ok {
		return l
	}
	return i18n.ES
}

func (f languageFile) info() LanguageInfo {
	choice := f.Choice
	if _, ok := i18n.Parse(choice); !ok {
		choice = autoLanguage
	}
	detected, ok := i18n.Parse(f.Detected)
	if !ok {
		detected = i18n.ES
	}
	return LanguageInfo{Choice: choice, Active: string(f.active()), Detected: string(detected)}
}

func (s *Service) initLanguage() {
	var f languageFile
	if b, err := os.ReadFile(languagePath()); err == nil {
		json.Unmarshal(b, &f)
	}
	if _, ok := i18n.Parse(f.Detected); !ok {
		l, sure := detectLanguage(windowsLanguage(), selfCountry)
		f.Detected, f.guess = string(l), !sure
		if sure {
			saveLanguage(f)
		}
	}
	s.lang = f
	i18n.Set(s.lang.active())
}

func saveLanguage(f languageFile) error {
	if f.guess {
		f.Detected = ""
	}
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	return os.WriteFile(languagePath(), b, 0o644)
}

func (s *Service) Language() LanguageInfo {
	s.langOnce.Do(s.initLanguage)
	s.langMu.Lock()
	defer s.langMu.Unlock()
	return s.lang.info()
}

func (s *Service) SetLanguage(choice string) (LanguageInfo, error) {
	s.langOnce.Do(s.initLanguage)
	if _, ok := i18n.Parse(choice); !ok && choice != autoLanguage {
		return LanguageInfo{}, errors.New(i18n.T("Toque no tiene ese idioma"))
	}
	if choice == autoLanguage {
		choice = ""
	}
	s.langMu.Lock()
	s.lang.Choice = choice
	i18n.Set(s.lang.active())
	f := s.lang
	s.langMu.Unlock()
	s.setUpdate(func(*UpdateInfo) {})
	saveLanguage(f)
	return f.info(), nil
}
