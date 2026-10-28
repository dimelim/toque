package app

import (
	"sort"

	"toque/internal/i18n"
	"toque/internal/winx"
)

func (s *Service) sessionInfo() SessionInfo {
	ss := s.sess
	ss.mu.Lock()
	defer ss.mu.Unlock()
	info := SessionInfo{
		Enabled: ss.cfg.Enabled,
		State:   ss.state,
		Prep:    ss.prep,
		Notes:   []SessionNote{},
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
	info.Actions = []SessionAction{ss.selfAction(), ss.windowsAction()}
	if laptop, _ := winx.Laptop(); laptop {
		info.Actions = append(info.Actions, ss.powerAction())
	}
	info.Actions = append(info.Actions, ss.backgroundAction(), ss.watchAction())
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
		Title:   i18n.T("Toque se aparta"),
		What:    i18n.T("Mientras juegas, la ventana de Toque deja de animarse y su interfaz pasa a modo eficiencia. Vuelve a la normalidad en cuanto la abres."),
		Tested:  true,
		Enabled: ss.cfg.Self,
		Status:  ss.status(ss.cfg.Self),
		Procs:   []SessionProc{},
	}
	if a.Status == "active" && ss.watch.samples > 0 {
		a.Detail = i18n.Tf("Toque usa un %s %% del procesador en esta partida.", decimal(share(ss.watch.own, ss.watch.samples)))
	}
	return a
}

func (ss *session) backgroundAction() SessionAction {
	a := SessionAction{
		ID:         "background",
		VerifyID:   backgroundID,
		Experiment: ss.experiment[backgroundID],
		Title:      i18n.T("Apps de fondo en modo eficiencia"),
		What:       i18n.T("Pone en modo eficiencia los programas conocidos que gastan CPU mientras juegas, como navegadores, launchers o fondos animados. Nunca toca los que están sonando, Discord, tu antivirus ni la app que tengas delante."),
		Tested:     false,
		Enabled:    ss.cfg.Background,
		Status:     ss.status(ss.cfg.Background),
		Procs:      []SessionProc{},
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
		a.Detail = i18n.T("Ninguna app de fondo conocida está gastando CPU ahora.")
	} else {
		a.Detail = i18n.Tf("%d en modo eficiencia ahora.", len(a.Procs))
	}
	return a
}

func (ss *session) windowsAction() SessionAction {
	a := SessionAction{
		ID:         "windows",
		VerifyID:   visualID,
		Experiment: ss.experiment[visualID],
		Title:      i18n.T("Windows en modo juego"),
		What:       i18n.T("Mientras juegas, Windows deja de grabar tus partidas en segundo plano y apaga animaciones, sombras y transparencia. Microsoft reconoce que esa grabación cuesta rendimiento salvo en gráficas de gama alta. Al cerrar Roblox todo vuelve a como lo tenías."),
		Tested:     false,
		Enabled:    ss.cfg.Windows,
		Status:     ss.status(ss.cfg.Windows),
		Procs:      []SessionProc{},
	}
	if a.Status == "active" {
		if visualChange.applied() {
			a.Detail = i18n.T("Apagado hasta que cierres Roblox.")
		} else {
			a.Detail = i18n.T("Tu Windows ya lo tenía todo apagado.")
		}
	}
	return a
}

func (ss *session) powerAction() SessionAction {
	a := SessionAction{
		ID:         "power",
		VerifyID:   powerID,
		Experiment: ss.experiment[powerID],
		Title:      i18n.T("Portátil en máximo rendimiento"),
		What:       i18n.T("Mientras juegas con el portátil enchufado, el modo de energía de Windows pasa a Máximo rendimiento y vuelve a tu ajuste al cerrar Roblox. Durante la partida el portátil gasta más y se calienta más."),
		Tested:     false,
		Enabled:    ss.cfg.Power,
		Status:     ss.status(ss.cfg.Power),
		Procs:      []SessionProc{},
	}
	if a.Status == "active" {
		switch {
		case powerChange.applied():
			a.Detail = i18n.T("Máximo rendimiento hasta que cierres Roblox.")
		case !powerApplicable():
			a.Detail = i18n.T("Solo funciona con el portátil enchufado y el plan de energía Equilibrado.")
		default:
			a.Detail = i18n.T("Tu Windows ya estaba en Máximo rendimiento.")
		}
	}
	return a
}

func (ss *session) watchAction() SessionAction {
	a := SessionAction{
		ID:      "watch",
		Title:   i18n.T("Vigilar qué te frena"),
		What:    i18n.T("Mira cada 10 segundos qué programas gastan procesador, gráfica y memoria y te lo cuenta aquí. No cambia nada."),
		Tested:  true,
		Enabled: ss.cfg.Watch,
		Status:  ss.status(ss.cfg.Watch),
		Procs:   []SessionProc{},
	}
	switch {
	case a.Status != "active":
	case ss.watch.samples == 1:
		a.Detail = i18n.T("1 revisión en esta partida.")
	default:
		a.Detail = i18n.Tf("%d revisiones en esta partida.", ss.watch.samples)
	}
	return a
}
