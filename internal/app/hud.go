package app

import (
	"encoding/json"
	"os"
	"path/filepath"

	"toque/internal/overlay"
	"toque/internal/winx"
)

func hudPath() string { return filepath.Join(winx.DataDir(), "hud.json") }

func loadHudConfig() overlay.Config {
	cfg := overlay.Defaults()
	if b, err := os.ReadFile(hudPath()); err == nil {
		json.Unmarshal(b, &cfg)
	}
	cfg.Move = false
	cfg.Demo = false
	return cfg
}

func (s *Service) HudConfig() overlay.Config {
	return s.hud.Config()
}

func (s *Service) SetHudConfig(cfg overlay.Config) overlay.Config {
	cfg.Demo = false
	s.hud.Apply(cfg)
	out := s.hud.Config()
	saved := out
	saved.Move = false
	if b, err := json.MarshalIndent(saved, "", "  "); err == nil {
		winx.AsUser(func() error { return os.WriteFile(hudPath(), b, 0o644) })
	}
	return out
}

func (s *Service) MoveHudTo(corner string) overlay.Config {
	cfg := s.hud.Config()
	sw, sh := overlay.ScreenSize()
	pw, ph := overlay.PanelSize(cfg)
	const m = 32
	switch corner {
	case "tr":
		cfg.X, cfg.Y = sw-pw-m, m
	case "bl":
		cfg.X, cfg.Y = m, sh-ph-m
	case "br":
		cfg.X, cfg.Y = sw-pw-m, sh-ph-m
	default:
		cfg.X, cfg.Y = m, m
	}
	return s.SetHudConfig(cfg)
}
