package app

import (
	"fmt"
	"sort"
)

func (s *Service) sessionInfo() SessionInfo {
	ss := s.sess
	ss.mu.Lock()
	defer ss.mu.Unlock()
	info := SessionInfo{
		Enabled:    ss.cfg.Enabled,
		State:      ss.state,
		VerifyID:   backgroundID,
		Prep:       ss.prep,
		Experiment: ss.experiment,
		Notes:      []SessionNote{},
	}
	if info.Prep == nil {
		info.Prep = []PrepCheck{}
	}
	if ss.state == stateActive {
		info.Since = ss.since
	}
	if ss.cfg.Watch {
		info.Notes = ss.watch.notes()
	}
	info.Actions = []SessionAction{ss.selfAction(), ss.backgroundAction(), ss.watchAction()}
	return info
}

func (ss *session) status(on bool) string {
	switch {
	case !on || ss.state == stateOff:
		return "off"
	case ss.state == stateActive:
		return "active"
	}
	return "waiting"
}

func (ss *session) selfAction() SessionAction {
	a := SessionAction{
		ID:      "self",
		Title:   "Toque se aparta",
		What:    "Mientras juegas, la ventana de Toque deja de animarse y su interfaz pasa a modo eficiencia. Vuelve a la normalidad en cuanto la abres.",
		Tested:  true,
		Enabled: ss.cfg.Self,
		Status:  ss.status(ss.cfg.Self),
		Procs:   []SessionProc{},
	}
	if a.Status == "active" && ss.watch.samples > 0 {
		a.Detail = "Toque usa un " + decimal(share(ss.watch.own, ss.watch.samples)) + " % del procesador en esta partida."
	}
	return a
}

func (ss *session) backgroundAction() SessionAction {
	a := SessionAction{
		ID:      "background",
		Title:   "Apps de fondo en modo eficiencia",
		What:    "Pone en modo eficiencia los programas conocidos que gastan CPU mientras juegas, como navegadores, launchers o fondos animados. Nunca toca los que están sonando, Discord, tu antivirus ni la app que tengas delante.",
		Tested:  false,
		Enabled: ss.cfg.Background,
		Status:  ss.status(ss.cfg.Background),
		Procs:   []SessionProc{},
	}
	if a.Status != "active" {
		return a
	}
	throttled := throttler().Throttled()
	byApp := map[string]float64{}
	for _, p := range ss.procs {
		if _, ok := throttled[p.Name]; ok {
			byApp[label(p.Name)] += p.Core
		}
	}
	for name, core := range byApp {
		a.Procs = append(a.Procs, SessionProc{Name: name, CPU: round1(share(core, 1))})
	}
	sort.Slice(a.Procs, func(i, j int) bool { return a.Procs[i].CPU > a.Procs[j].CPU })
	if len(a.Procs) == 0 {
		a.Detail = "Ninguna app de fondo conocida está gastando CPU ahora."
	} else {
		a.Detail = fmt.Sprintf("%d en modo eficiencia ahora.", len(a.Procs))
	}
	return a
}

func (ss *session) watchAction() SessionAction {
	a := SessionAction{
		ID:      "watch",
		Title:   "Vigilar qué te frena",
		What:    "Mira cada 10 segundos qué programas gastan procesador y memoria y te lo cuenta aquí. No cambia nada.",
		Tested:  true,
		Enabled: ss.cfg.Watch,
		Status:  ss.status(ss.cfg.Watch),
		Procs:   []SessionProc{},
	}
	switch {
	case a.Status != "active":
	case ss.watch.samples == 1:
		a.Detail = "1 revisión en esta partida."
	default:
		a.Detail = fmt.Sprintf("%d revisiones en esta partida.", ss.watch.samples)
	}
	return a
}
