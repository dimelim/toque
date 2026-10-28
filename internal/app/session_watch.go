package app

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"toque/internal/eco"
	"toque/internal/sysmon"
)

const (
	lowRAMMB     = 1500
	topCPUShown  = 3
	minCPUShown  = 1.0
	minGPUShown  = 2.0
	defenderBusy = 3.0
)

var quietProcs = map[string]bool{
	"system": true, "registry": true, "memory compression": true, "secure system": true, "idle": true,
	"csrss.exe": true, "smss.exe": true, "wininit.exe": true, "services.exe": true, "lsass.exe": true,
	"winlogon.exe": true, "fontdrvhost.exe": true, "svchost.exe": true, "dwm.exe": true, "audiodg.exe": true,
	"robloxplayerbeta.exe": true, "robloxcrashhandler.exe": true,
}

type watchLog struct {
	samples int
	own     float64
	core    map[string]float64
	gpu     map[string]float64
	ram     map[string]float64
	seen    map[string]bool
	ramFree float64
}

func (w *watchLog) add(procs []sysmon.ProcCPU, own map[uint32]bool, gpu map[uint32]float64) {
	if w.core == nil {
		w.core, w.gpu, w.seen = map[string]float64{}, map[string]float64{}, map[string]bool{}
	}
	w.samples++
	ram := map[string]float64{}
	for _, p := range procs {
		if own[p.PID] {
			w.own += p.Core
			continue
		}
		w.seen[p.Name] = true
		if quietProcs[p.Name] {
			continue
		}
		w.core[p.Name] += p.Core
		w.gpu[p.Name] += gpu[p.PID]
		ram[p.Name] += p.RAMMB
	}
	w.ram = ram
	w.ramFree = sysmon.AvailableMB()
}

func share(core float64, samples int) float64 {
	if samples == 0 {
		return 0
	}
	return core / float64(samples) / float64(runtime.NumCPU())
}

func label(exe string) string {
	if l, ok := eco.Label(exe); ok {
		return l
	}
	return strings.TrimSuffix(exe, ".exe")
}

func (w watchLog) notes() []SessionNote {
	if w.samples == 0 {
		return []SessionNote{}
	}
	notes := []SessionNote{}
	if top := w.topCPU(); len(top) > 0 {
		n := SessionNote{ID: "cpu", Tone: "info", Title: "Lo que más CPU gastó mientras jugabas", Detail: "Media de toda la partida, en porcentaje del procesador entero.", Procs: top}
		if top[0].CPU >= 5 {
			n.Tone = "warn"
		}
		notes = append(notes, n)
	}
	if top := w.topGPU(); len(top) > 0 {
		n := SessionNote{ID: "gpu", Tone: "info", Title: "Lo que más gráfica gastó mientras jugabas", Detail: "Media de toda la partida, en porcentaje de la gráfica. Lo que la usa de fondo se la quita a Roblox.", Procs: top}
		if top[0].GPU >= 10 {
			n.Tone = "warn"
		}
		notes = append(notes, n)
	}
	if w.ramFree > 0 && w.ramFree < lowRAMMB {
		notes = append(notes, SessionNote{ID: "ram", Tone: "warn", Title: "Te queda poca memoria libre", Detail: fmt.Sprintf("Quedan %.0f MB libres. Cerrar alguno de estos programas antes de jugar libera memoria para Roblox.", w.ramFree), Procs: w.topRAM()})
	}
	if d := share(w.core["msmpeng.exe"], w.samples); d >= defenderBusy {
		notes = append(notes, SessionNote{ID: "defender", Tone: "info", Title: "Microsoft Defender estuvo analizando", Detail: "Gastó un " + decimal(d) + " % del procesador durante tu partida. Suele pasar después de descargas o actualizaciones de Windows.", Procs: []SessionProc{}})
	}
	if w.seen["exitlag.exe"] {
		notes = append(notes, SessionNote{ID: "exitlag", Tone: "info", Title: "ExitLag está abierto", Detail: "Cambia la ruta de tu conexión, así que el ping que ve Toque es el del túnel. Si Roblox te saca de la partida, prueba a jugar sin él.", Procs: []SessionProc{}})
	}
	return notes
}

func (w watchLog) byApp() []SessionProc {
	apps := map[string]*SessionProc{}
	for exe := range w.seen {
		if quietProcs[exe] || exe == "msmpeng.exe" {
			continue
		}
		name := label(exe)
		p, ok := apps[name]
		if !ok {
			p = &SessionProc{Name: name}
			apps[name] = p
		}
		p.CPU += share(w.core[exe], w.samples)
		p.GPU += w.gpu[exe] / float64(w.samples)
		p.RAMMB += w.ram[exe]
	}
	out := make([]SessionProc, 0, len(apps))
	for _, p := range apps {
		out = append(out, SessionProc{Name: p.Name, CPU: round1(p.CPU), GPU: round1(p.GPU), RAMMB: round1(p.RAMMB)})
	}
	return out
}

func top(list []SessionProc, by func(SessionProc) float64, min float64) []SessionProc {
	sort.Slice(list, func(i, j int) bool { return by(list[i]) > by(list[j]) })
	out := []SessionProc{}
	for _, p := range list {
		if len(out) == topCPUShown || by(p) < min {
			break
		}
		out = append(out, p)
	}
	return out
}

func (w watchLog) topCPU() []SessionProc {
	return top(w.byApp(), func(p SessionProc) float64 { return p.CPU }, minCPUShown)
}

func (w watchLog) topGPU() []SessionProc {
	return top(w.byApp(), func(p SessionProc) float64 { return p.GPU }, minGPUShown)
}

func (w watchLog) topRAM() []SessionProc {
	return top(w.byApp(), func(p SessionProc) float64 { return p.RAMMB }, 1)
}

func decimal(v float64) string {
	return strings.Replace(strconv.FormatFloat(v, 'f', 1, 64), ".", ",", 1)
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
