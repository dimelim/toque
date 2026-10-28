package app

import (
	"errors"
	"fmt"

	"toque/internal/catalog"
	"toque/internal/roblox"
	"toque/internal/tweaks"
)

var framePresets = []int{60, 120, 144, 240}

func framePreset(hz int) int {
	for _, p := range framePresets {
		if p >= hz {
			return p
		}
	}
	return framePresets[len(framePresets)-1]
}

func refreshHz() int {
	systemOnce.Do(func() { systemInfo = ReadSystem() })
	return systemInfo.RefreshHz
}

func (s *Service) prepChecks() []PrepCheck {
	out := []PrepCheck{}
	if c, ok := fpsCheck(); ok {
		out = append(out, c)
	}
	return append(out, s.provenCheck())
}

func fpsCheck() (PrepCheck, bool) {
	inst, err := roblox.Find()
	if err != nil {
		return PrepCheck{}, false
	}
	st, err := roblox.LoadBasicSettings(inst.SettingsPath())
	if err != nil {
		return PrepCheck{}, false
	}
	limit, _ := st.Int("FramerateCap")
	hz := refreshHz()
	c := PrepCheck{ID: "fps-cap", Title: "Límite de FPS"}
	switch {
	case hz <= 60 || limit >= hz:
		c.Done = true
		c.Detail = fmt.Sprintf("Roblox puede llegar a los %d Hz de tu pantalla.", hz)
		return c, true
	case limit <= 0:
		c.Detail = fmt.Sprintf("Roblox usa su límite por defecto, que es de 60 FPS, y tu pantalla es de %d Hz.", hz)
	default:
		c.Detail = fmt.Sprintf("Roblox está limitado a %d FPS y tu pantalla es de %d Hz.", limit, hz)
	}
	c.Action = fmt.Sprintf("Subir a %d FPS", framePreset(hz))
	if len(roblox.Running()) > 0 {
		c.Blocked = "Cierra Roblox para cambiarlo."
	}
	return c, true
}

func (s *Service) provenCheck() PrepCheck {
	total, applied := 0, 0
	for _, o := range s.Optimizations() {
		if (o.Verdict != catalog.VerdictProven && o.Verdict != catalog.VerdictTradeoff) || !o.Applicable {
			continue
		}
		total++
		if o.Applied {
			applied++
		}
	}
	c := PrepCheck{ID: "proven", Title: "Ajustes probados", Done: total > 0 && applied == total}
	c.Detail = fmt.Sprintf("Tienes %d de %d aplicados.", applied, total)
	if !c.Done {
		c.Action = "Ver ajustes"
	}
	return c
}

func (s *Service) ApplyPrep(id string) (SessionInfo, error) {
	if id != "fps-cap" {
		return s.Session(), errors.New("no conozco ese ajuste")
	}
	if err := s.idle(); err != nil {
		return s.Session(), err
	}
	if len(roblox.Running()) > 0 {
		return s.Session(), errors.New("cierra Roblox para cambiar el límite de FPS")
	}
	t, err := tweaks.Get(fmt.Sprintf("setting:FramerateCap=%d", framePreset(refreshHz())))
	if err != nil {
		return s.Session(), err
	}
	if err := tweaks.Apply(s.context(), t); err != nil {
		return s.Session(), err
	}
	return s.Session(), nil
}
