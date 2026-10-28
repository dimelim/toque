package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"toque/internal/catalog"
	"toque/internal/i18n"
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
	if c, ok := gameModeCheck(); ok {
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

func gameModeCheck() (PrepCheck, bool) {
	t := tweaks.GameMode()
	managed, _ := tweaks.Entries()
	_, mine := managed[t.ID()]
	switch {
	case tweaks.GameModeOff():
		return PrepCheck{ID: "gamemode", Title: i18n.T("Modo de juego de Windows"), Detail: i18n.T("Lo tienes apagado. Windows lo trae activado porque evita que instale controladores y te muestre notificaciones mientras juegas."), Action: i18n.T("Activar")}, true
	case mine:
		return PrepCheck{ID: "gamemode", Title: i18n.T("Modo de juego de Windows"), Detail: i18n.T("Activado de nuevo por Toque."), Done: true, Action: i18n.T("Quitar")}, true
	}
	return PrepCheck{}, false
}

func windowedCheck() (PrepCheck, bool) {
	t := tweaks.WindowedGames()
	st := t.State(tweaks.Context{})
	managed, _ := tweaks.Entries()
	_, mine := managed[t.ID()]
	if !st.Applicable || (!mine && !strings.Contains(lastPresentMode(), "Copy with")) {
		return PrepCheck{}, false
	}
	c := PrepCheck{ID: "windowed", Title: i18n.T("Juegos en ventana")}
	if st.Applied {
		c.Done, c.Detail = true, i18n.T("Activado. Reabre Roblox y mide de nuevo para comprobar que ya no copia cada frame.")
		if mine {
			c.Action = i18n.T("Quitar")
		} else {
			c.Detail = i18n.T("Ya lo tenías activado en Windows.")
		}
		return c, true
	}
	c.Detail = i18n.T("En tu última medición Roblox se mostraba copiando cada frame al escritorio. Windows 11 puede pasarlo al modo flip, que Microsoft documenta con menos retraso. Se nota al reabrir Roblox.")
	c.Action = i18n.T("Activar")
	return c, true
}

func windowsCheck() PrepCheck {
	c := PrepCheck{ID: "windows", Title: i18n.T("Windows más limpio"), Action: i18n.T("Ver ajustes")}
	pending, unknown := pendingPrivacy()
	switch {
	case pending == 1:
		c.Detail = i18n.T("Te queda 1 ajuste de privacidad por quitar.")
	case pending > 1:
		c.Detail = i18n.Tf("Te quedan %d ajustes de privacidad por quitar.", pending)
	case unknown > 0:
		c.Detail = i18n.T("No pude comprobar algunos ajustes de Windows.")
	default:
		c.Done, c.Action, c.Detail = true, "", i18n.T("Sin telemetría opcional, publicidad ni sugerencias.")
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
	c := PrepCheck{ID: "fps-cap", Title: i18n.T("Límite de FPS")}
	switch {
	case hz <= 60 || limit >= hz:
		c.Done = true
		c.Detail = i18n.Tf("Roblox puede llegar a los %d Hz de tu pantalla.", hz)
		return c, true
	case limit <= 0:
		c.Detail = i18n.Tf("Roblox usa su límite por defecto, que es de 60 FPS, y tu pantalla es de %d Hz.", hz)
	default:
		c.Detail = i18n.Tf("Roblox está limitado a %d FPS y tu pantalla es de %d Hz.", limit, hz)
	}
	c.Action = i18n.Tf("Subir a %d FPS", framePreset(hz))
	if len(roblox.Running()) > 0 {
		c.Blocked = i18n.T("Cierra Roblox para cambiarlo.")
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
	c := PrepCheck{ID: "proven", Title: i18n.T("Ajustes probados"), Done: total > 0 && applied == total}
	c.Detail = i18n.Tf("Tienes %d de %d aplicados.", applied, total)
	if !c.Done {
		c.Action = i18n.T("Ver ajustes")
	}
	return c
}

func (s *Service) ApplyPrep(id string) (SessionInfo, error) {
	if err := s.idle(); err != nil {
		return s.Session(), err
	}
	switch id {
	case "gamemode":
		t := tweaks.GameMode()
		if managed, _ := tweaks.Entries(); managed != nil {
			if _, mine := managed[t.ID()]; mine {
				return s.Session(), tweaks.Revert(s.context(), t)
			}
		}
		return s.Session(), tweaks.Apply(s.context(), t)
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
		return s.Session(), errors.New(i18n.T("no conozco ese ajuste"))
	}
	if len(roblox.Running()) > 0 {
		return s.Session(), errors.New(i18n.T("cierra Roblox para cambiar el límite de FPS"))
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
