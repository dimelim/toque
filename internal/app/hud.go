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
	saveHudConfig(out)
	return out
}

func saveHudConfig(cfg overlay.Config) {
	cfg.Move = false
	if b, err := json.MarshalIndent(cfg, "", "  "); err == nil {
		winx.AsUser(func() error { return os.WriteFile(hudPath(), b, 0o644) })
	}
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

type HudLayout struct {
	ScreenW int `json:"screenW"`
	ScreenH int `json:"screenH"`
	PanelW  int `json:"panelW"`
	PanelH  int `json:"panelH"`
}

func (s *Service) HudLayout() HudLayout {
	sw, sh := overlay.ScreenSize()
	pw, ph := overlay.PanelSize(s.hud.Config())
	return HudLayout{ScreenW: sw, ScreenH: sh, PanelW: pw, PanelH: ph}
}

func (s *Service) PlaceHud(x, y int, save bool) overlay.Config {
	l := s.HudLayout()
	x = max(0, min(x, l.ScreenW-l.PanelW))
	y = max(0, min(y, l.ScreenH-l.PanelH))
	cfg := s.hud.MoveTo(x, y)
	if save {
		saveHudConfig(cfg)
	}
	return cfg
}

func (s *Service) PreviewHud(on bool) {
	s.hud.SetPreview(on)
}
