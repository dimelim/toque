package app

import (
	"os"
	"path/filepath"
	"strings"

	"toque/internal/winx"
)

type WhatsNew struct {
	Version string `json:"version"`
	Show    bool   `json:"show"`
}

func versionPath() string { return filepath.Join(winx.DataDir(), "version.txt") }

func returningUser() bool {
	for _, name := range []string{"update", "journal.json", "history", "dcmap.json", "places.json"} {
		if _, err := os.Stat(filepath.Join(winx.DataDir(), name)); err == nil {
			return true
		}
	}
	return false
}

func (s *Service) WhatsNew() WhatsNew {
	prev, _ := os.ReadFile(versionPath())
	last := strings.TrimSpace(string(prev))
	if last == Version {
		return WhatsNew{Version: Version}
	}
	if last == "" && !returningUser() {
		s.SeenWhatsNew()
		return WhatsNew{Version: Version}
	}
	return WhatsNew{Version: Version, Show: true}
}

func (s *Service) SeenWhatsNew() {
	os.WriteFile(versionPath(), []byte(Version), 0o644)
}
