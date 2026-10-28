package app

import (
	"os"
	"path/filepath"
	"strings"
)

type WhatsNew struct {
	Version string `json:"version"`
	Show    bool   `json:"show"`
}

func versionPath() string { return filepath.Join(systemDir(), "version.txt") }

func returningUser() bool {
	for _, name := range []string{"journal.json", "datacenters.json", "places.json", "session.json", "hud.json"} {
		if _, err := os.Stat(filepath.Join(systemDir(), name)); err == nil {
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
	if last == "" && !s.returning {
		s.SeenWhatsNew()
		return WhatsNew{Version: Version}
	}
	return WhatsNew{Version: Version, Show: true}
}

func (s *Service) SeenWhatsNew() {
	os.WriteFile(versionPath(), []byte(Version), 0o644)
}
