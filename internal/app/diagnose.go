package app

import (
	"fmt"
	"net/netip"
	"path/filepath"
	"strings"
	"time"

	"toque/internal/bench"
	"toque/internal/netq"
	"toque/internal/roblox"
	"toque/internal/sysmon"
	"toque/internal/tweaks"
	"toque/internal/winx"
)

type Finding struct {
	ID     string `json:"id"`
	Level  string `json:"level"`
	Area   string `json:"area"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Fix    string `json:"fix,omitempty"`
}

type PathStep struct {
	Label  string  `json:"label"`
	Addr   string  `json:"addr"`
	RTT    float64 `json:"rttMs"`
	Jitter float64 `json:"jitterMs"`
	Loss   float64 `json:"lossPct"`
	Kind   string  `json:"kind"`
}

type Diagnosis struct {
	At           time.Time           `json:"at"`
	System       SystemInfo          `json:"system"`
	Path         []PathStep          `json:"path"`
	Game         *bench.Window       `json:"game"`
	Pagefile     tweaks.PagefileInfo `json:"pagefile"`
	PowerPlan    string              `json:"powerPlan"`
	PowerName    string              `json:"powerName"`
	RAMAvailMB   float64             `json:"ramAvailMb"`
	Hogs         []sysmon.ProcMem    `json:"hogs"`
	FramerateCap int                 `json:"framerateCap"`
	Findings     []Finding           `json:"findings"`
}

var knownPlans = map[string]string{
	tweaks.PlanBalanced:                    "Equilibrado",
	tweaks.PlanHigh:                        "Alto rendimiento",
	"a1841308-3541-4fab-bc81-f71556f20b4a": "Economizador",
	"e9a42b02-d5df-448d-aa00-03f14749eb61": "Máximo rendimiento",
}

func lastServerIP(inst roblox.Install) string {
	log := roblox.NewestPlayerLog(inst.LogDir(), time.Now().Add(-14*24*time.Hour))
	ip := ""
	for _, ev := range roblox.ScanFile(log) {
		if ev.Kind == roblox.EventUdmux {
			ip = ev.IP
		}
	}
	return ip
}

func networkPath(target string) []PathStep {
	addr, err := netip.ParseAddr(target)
	if err != nil {
		addr = netip.MustParseAddr("128.116.0.1")
	}
	var out []PathStep
	seenPublic := false
	for _, h := range netq.Trace(addr, 10, 4) {
		step := PathStep{Addr: h.Addr, RTT: h.RTT, Jitter: h.Jitter, Loss: h.Loss}
		switch {
		case h.Private && len(out) == 0:
			step.Label, step.Kind = "Router", "local"
		case h.Private:
			step.Label, step.Kind = "Módem / red local", "local"
		case !seenPublic:
			step.Label, step.Kind = "Tu proveedor", "isp"
			seenPublic = true
		default:
			step.Label, step.Kind = "Red del proveedor", "isp"
		}
		out = append(out, step)
	}
	return out
}

func (s *Service) Diagnose(withGame bool) (Diagnosis, error) {
	ctx, done, err := s.begin("diagnose")
	if err != nil {
		return Diagnosis{}, err
	}
	defer done()
	d := Diagnosis{At: time.Now(), System: ReadSystem()}
	systemOnce.Do(func() { systemInfo = d.System })
	s.progress("diagnose", "system", 0.05, "Revisando memoria, disco y energía")
	d.Pagefile = tweaks.Pagefile()
	d.RAMAvailMB = sysmon.AvailableMB()
	if plan, err := tweaks.ActivePlan(); err == nil {
		d.PowerPlan = plan
		d.PowerName = tweaks.PlanName(plan)
	}
	procs, _ := winx.Processes()
	list := make([]sysmon.Proc, 0, len(procs))
	for _, p := range procs {
		list = append(list, sysmon.Proc{PID: p.PID, Name: p.Name})
	}
	d.Hogs = sysmon.TopMemory(list, 6, func(n string) bool { return strings.HasPrefix(n, "robloxplayer") || strings.HasPrefix(n, "toque") })

	inst, instErr := roblox.Find()
	if instErr == nil {
		if st, err := roblox.LoadBasicSettings(inst.SettingsPath()); err == nil {
			d.FramerateCap, _ = st.Int("FramerateCap")
		}
	}

	target := ""
	sess, sessErr := roblox.Session{}, fmt.Errorf("no session")
	if instErr == nil {
		sess, sessErr = currentSession(inst)
		if sessErr == nil {
			target = sess.ServerIP
		} else {
			target = lastServerIP(inst)
		}
	}
	s.progress("diagnose", "path", 0.15, "Midiendo la ruta hasta los servidores de Roblox")
	d.Path = networkPath(target)

	launched := false
	if instErr == nil && sessErr != nil && withGame {
		s.progress("diagnose", "join", 0.4, "Abriendo Realistic Street Soccer para medir dentro del juego")
		js, err := roblox.Join(ctx, inst, roblox.RealisticStreetSoccer, "", 90*time.Second)
		if err == nil {
			sess, sessErr, launched = js, nil, true
			select {
			case <-ctx.Done():
			case <-time.After(12 * time.Second):
			}
		}
	}
	if sessErr == nil && ctx.Err() == nil {
		env, err := s.env()
		if err == nil {
			s.progress("diagnose", "game", 0.6, "Midiendo frames y ping real durante 20 s")
			winx.Focus(winx.MainWindow(sess.PID))
			w := bench.Measure(ctx, env, sess, 20*time.Second, filepath.Join(winx.DataDir("diagnostics"), time.Now().Format("20060102-150405")), "diagnostico")
			d.Game = &w
			if w.Net.Confident {
				d.Path = append(d.Path, PathStep{Label: "Servidor de Roblox", Addr: sess.ServerIP, RTT: w.Net.RTT, Jitter: w.Net.Jitter, Loss: w.Net.LossDownPct, Kind: "game"})
				rememberDatacenter(sess.ServerIP, w.Net.RTT)
			}
		}
	}
	if launched {
		roblox.CloseAll(10 * time.Second)
	}
	d.Findings = findings(d)
	return d, nil
}

func dec(v float64) string {
	return strings.Replace(fmt.Sprintf("%.1f", v), ".", ",", 1)
}

func findings(d Diagnosis) []Finding {
	var f []Finding
	add := func(id, level, area, title, detail, fix string) {
		f = append(f, Finding{ID: id, Level: level, Area: area, Title: title, Detail: detail, Fix: fix})
	}
	pf := d.Pagefile
	if pf.OnHDD {
		add("pagefile-hdd", "bad", "memoria", "El archivo de paginación está en un disco mecánico",
			fmt.Sprintf("Windows lo puso en %s. Con %.0f GB de RAM, cada vez que el juego necesita una página guardada ahí espera entre 9 y 70 ms, y eso se nota como tirón.", strings.Join(pf.Drives, ", "), d.System.RAMGB),
			"pagefile-ssd")
	} else if len(pf.Drives) > 0 {
		add("pagefile-ok", "ok", "memoria", "Archivo de paginación en SSD", "Las páginas que Windows recupera durante la partida se leen en alrededor de 1 ms.", "")
	}
	if d.System.RAMGB <= 8.5 && d.RAMAvailMB < 2048 {
		names := []string{}
		for i, h := range d.Hogs {
			if i >= 3 {
				break
			}
			names = append(names, fmt.Sprintf("%s (%.0f MB)", h.Name, h.MB))
		}
		add("ram-low", "warn", "memoria", "Poca RAM libre para jugar",
			fmt.Sprintf("Quedan %.0f MB libres y Roblox usa cerca de 2 GB. Lo que más ocupa ahora es %s. Cerrarlo antes de jugar reduce los tirones.", d.RAMAvailMB, strings.Join(names, ", ")), "")
	}
	planName := d.PowerName
	if planName == "" {
		planName = knownPlans[d.PowerPlan]
	}
	if planName != "" {
		add("power-plan", "info", "sistema", "Plan de energía "+planName, "En las pruebas, cambiar entre Equilibrado, Alto rendimiento y un plan personalizado no movió los FPS ni los tirones en un PC de escritorio, así que Toque no lo toca.", "")
	}
	if d.FramerateCap > 0 && d.FramerateCap < d.System.RefreshHz {
		add("fps-cap", "warn", "fps", fmt.Sprintf("Roblox limitado a %d FPS con un monitor de %d Hz", d.FramerateCap, d.System.RefreshHz),
			"Cada frame tarda más en llegar a la pantalla de lo que tu monitor permite. Sube el límite en la configuración de Roblox.", "")
	}
	bestDC := 0.0
	for _, e := range loadDatacenters() {
		if e.RTT > 0 && (bestDC == 0 || e.RTT < bestDC) {
			bestDC = e.RTT
		}
	}
	for _, p := range d.Path {
		switch {
		case p.Kind == "local" && p.Label == "Router" && p.RTT > 4:
			add("lan", "warn", "red", "Tu red local añade retraso", fmt.Sprintf("Llegar al router cuesta %s ms. Por cable suele ser menos de 1 ms, así que el Wi-Fi o el router están sumando retraso.", dec(p.RTT)), "")
		case p.Kind == "local" && p.Loss > 2:
			add("lan-loss", "warn", "red", "Se pierden paquetes antes de salir de casa", fmt.Sprintf("El %s no respondió al %s %% de las pruebas.", strings.ToLower(p.Label), dec(p.Loss)), "")
		case p.Kind == "game" && (p.RTT > 100 || (bestDC > 0 && p.RTT > bestDC+25)):
			detail := "Buscar otro servidor de Realistic Street Soccer suele bajar mucho el ping."
			if bestDC > 0 && p.RTT > bestDC+25 {
				detail = fmt.Sprintf("Ya te medí un centro de datos de Roblox a %.0f ms. Buscar servidor puede dejarte ahí.", bestDC)
			}
			add("far-server", "warn", "red", fmt.Sprintf("Servidor lejano, %.0f ms de ping", p.RTT), detail, "server")
		case p.Kind == "game" && p.Loss > 1:
			add("loss", "bad", "red", fmt.Sprintf("Se pierden paquetes del servidor (%s %%)", dec(p.Loss)), "Eso se nota como saltos del balón y de los jugadores.", "")
		}
	}
	if g := d.Game; g != nil && g.Frames.Frames > 0 {
		if !strings.Contains(g.Frames.PresentMode, "Independent Flip") && !strings.Contains(g.Frames.PresentMode, "Legacy Flip") {
			add("present", "warn", "fps", "Roblox pasa por el compositor de Windows", "El modo "+g.Frames.PresentMode+" suma un frame de retraso. Pon Roblox en pantalla completa.", "")
		} else {
			add("present-ok", "ok", "fps", "Presentación directa a pantalla", "Roblox dibuja sin pasar por el compositor de Windows, el camino de menor retraso.", "")
		}
		if g.Frames.StutterPerMin > 15 {
			add("stutter", "warn", "fps", fmt.Sprintf("%.0f tirones por minuto", g.Frames.StutterPerMin), fmt.Sprintf("El 1 %% de los peores frames va a %.0f FPS aunque la media sea %.0f.", g.Frames.Low1FPS, g.Frames.AvgFPS), "")
		}
		if g.Sys.HardFaults > 150 {
			add("paging", "warn", "memoria", "Windows lee memoria del disco durante la partida", fmt.Sprintf("%.0f lecturas de paginación por segundo de media.", g.Sys.HardFaults), "")
		}
	}
	if len(f) == 0 {
		add("all-ok", "ok", "sistema", "Todo en orden", "No encontré nada que esté frenando a Roblox en este equipo.", "")
	}
	return f
}
