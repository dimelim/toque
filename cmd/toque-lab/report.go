package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"toque/internal/bench"
	"toque/internal/winx"
)

var reportMetrics = []struct {
	key, label, unit string
	digits           int
}{
	{"low1", "1 % low", "FPS", 1},
	{"low01", "0,1 % low", "FPS", 1},
	{"stutterMin", "Tirones/min", "", 1},
	{"ftStdev", "Variación frametime", "ms", 2},
	{"latMs", "Latencia de render", "ms", 2},
	{"fps", "FPS medio", "FPS", 1},
	{"rtt", "Ping real", "ms", 1},
	{"ramRoblox", "RAM de Roblox", "MB", 0},
	{"hardFaults", "Lecturas de paginación", "/s", 0},
	{"gpuBusy", "GPU por frame", "ms", 2},
}

func cmdReport(args []string) error {
	out := "docs/BENCHMARKS.md"
	if len(args) > 0 {
		out = args[0]
	}
	files, _ := filepath.Glob(filepath.Join(winx.DataDir("experiments"), "*", "result.json"))
	sort.Strings(files)
	var b strings.Builder
	b.WriteString("# Resultados de los experimentos A/B\n\n")
	b.WriteString("Cada experimento alterna sin el cambio (A) y con el cambio (B) en orden ABBA dentro de Realistic Street Soccer. ")
	b.WriteString("Se compara la mediana de cada métrica por ventana, con intervalo de confianza bootstrap al 95 % y prueba de Mann-Whitney. ")
	b.WriteString("Una diferencia cuenta como real solo si el intervalo no cruza el cero y p < 0,05.\n\n")
	b.WriteString("\n\n")
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var r bench.ABResult
		if json.Unmarshal(raw, &r) != nil {
			continue
		}
		r.Compare, r.Verdict, r.Summary = bench.Evaluate(r.Samples, r.Config.Primary)
		valid := 0
		for _, s := range r.Samples {
			if s.Valid {
				valid++
			}
		}
		fmt.Fprintf(&b, "## %s\n\n", r.Config.Tweak)
		fmt.Fprintf(&b, "%s · %d ciclos · ventanas de %s · %d válidas de %d", r.Started.Format("2006-01-02 15:04"), r.Config.Cycles, r.Config.Window, valid, len(r.Samples))
		if r.Config.Stress != "" {
			fmt.Fprintf(&b, " · carga de fondo %s", r.Config.Stress)
		}
		fmt.Fprintf(&b, "\n\nVeredicto **%s**. %s.\n\n", r.Verdict, r.Summary)
		b.WriteString("| Métrica | A (sin) | B (con) | Cambio | IC 95 % | p | Concluyente |\n|---|---|---|---|---|---|---|\n")
		for _, m := range reportMetrics {
			c, ok := r.Compare[m.key]
			if !ok {
				continue
			}
			mark := "no"
			if c.Conclusive {
				mark = "**sí**"
			}
			fmt.Fprintf(&b, "| %s | %.*f %s | %.*f %s | %+.1f %% | %+.1f %% a %+.1f %% | %.3f | %s |\n",
				m.label, m.digits, c.MedianA, m.unit, m.digits, c.MedianB, m.unit, c.DeltaPct, c.CILowPct, c.CIHighPct, c.PValue, mark)
		}
		b.WriteString("\n")
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		return err
	}
	fmt.Println("wrote", out, len(files), "experiments")
	return nil
}
