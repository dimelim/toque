package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

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
	if c, ok := windowedCheck(); ok {
		out = append(out, c)
	}
	return append(out, s.provenCheck(), windowsCheck())
}

func lastPresentMode() string {
	mode, at := "", time.Time{}
	if d := loadLast[Diagnosis]("diagnosis.json"); d != nil && d.Game != nil {
		mode, at = d.Game.Frames.PresentMode, d.At
	}
	if r := latestRun(); r != nil && r.At.After(at) {
		mode = r.Window.Frames.PresentMode
	}
	return mode
}

func windowedCheck() (PrepCheck, bool) {
	t := tweaks.WindowedGames()
	st := t.State(tweaks.Context{})
	managed, _ := tweaks.Entries()
	_, mine := managed[t.ID()]
	if !st.Applicable || (!mine && !strings.Contains(lastPresentMode(), "Copy with")) {
		return PrepCheck{}, false
	}
	c := PrepCheck{ID: "windowed", Title: "Juegos en ventana"}
	if st.Applied {
		c.Done, c.Detail = true, "Activado. Reabre Roblox y mide de nuevo para comprobar que ya no copia cada frame."
		if mine {
			c.Action = "Quitar"
		} else {
			c.Detail = "Ya lo tenías activado en Windows."
		}
		return c, true
	}
	c.Detail = "En tu última medición Roblox se mostraba copiando cada frame al escritorio. Windows 11 puede pasarlo al modo flip, que Microsoft documenta con menos retraso. Se nota al reabrir Roblox."
	c.Action = "Activar"
	return c, true
}

func windowsCheck() PrepCheck {
	c := PrepCheck{ID: "windows", Title: "Windows más limpio", Action: "Ver ajustes"}
	pending, unknown := pendingPrivacy()
	switch {
	case pending == 1:
		c.Detail = "Te queda 1 ajuste de privacidad por quitar."
	case pending > 1:
		c.Detail = fmt.Sprintf("Te quedan %d ajustes de privacidad por quitar.", pending)
	case unknown > 0:
		c.Detail = "No pude comprobar algunos ajustes de Windows."
	default:
		c.Done, c.Action, c.Detail = true, "", "Sin telemetría opcional, publicidad ni sugerencias."
	}
	return c
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
	if err := s.idle(); err != nil {
		return s.Session(), err
	}
	switch id {
	case "windowed":
		t := tweaks.WindowedGames()
		if managed, _ := tweaks.Entries(); managed != nil {
			if _, mine := managed[t.ID()]; mine && t.State(s.context()).Applied {
				return s.Session(), tweaks.Revert(s.context(), t)
			}
		}
		return s.Session(), tweaks.Apply(s.context(), t)
	case "fps-cap":
	default:
		return s.Session(), errors.New("no conozco ese ajuste")
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
