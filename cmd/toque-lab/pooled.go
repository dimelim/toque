package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"toque/internal/bench"
	"toque/internal/catalog"
	"toque/internal/stats"
)

type pool struct {
	tweak, stress string
	runs          int
	a, b          map[string][]float64
	primary       map[string]bool
}

var verdictLabel = map[string]string{
	catalog.VerdictProven:      "Probada",
	catalog.VerdictTradeoff:    "Probada, con coste visual",
	catalog.VerdictSituational: "Según tu caso",
	catalog.VerdictPending:     "En estudio",
	catalog.VerdictRejected:    "Descartada",
}

func summary() string {
	var b strings.Builder
	b.WriteString("## Resumen\n\n| Optimización | Veredicto final | Evidencia |\n|---|---|---|\n")
	order := []string{catalog.VerdictProven, catalog.VerdictTradeoff, catalog.VerdictSituational, catalog.VerdictPending, catalog.VerdictRejected}
	for _, v := range order {
		for _, e := range catalog.Catalog {
			if e.Verdict == v {
				fmt.Fprintf(&b, "| %s | %s | %s |\n", e.Title, verdictLabel[v], e.Lab)
			}
		}
	}
	b.WriteString("\n")
	return b.String()
}

func es(v float64, digits int) string {
	return strings.Replace(fmt.Sprintf("%.*f", digits, v), ".", ",", 1)
}

func esSigned(v float64, digits int) string {
	return strings.Replace(fmt.Sprintf("%+.*f", digits, v), ".", ",", 1)
}

func pooled(files []string) string {
	groups := map[string]*pool{}
	var order []string
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var r bench.ABResult
		if json.Unmarshal(raw, &r) != nil {
			continue
		}
		key := r.Config.Tweak + "|" + r.Config.Stress
		g, ok := groups[key]
		if !ok {
			g = &pool{tweak: r.Config.Tweak, stress: r.Config.Stress, a: map[string][]float64{}, b: map[string][]float64{}, primary: map[string]bool{}}
			groups[key] = g
			order = append(order, key)
		}
		g.runs++
		for _, p := range r.Config.Primary {
			g.primary[p] = true
		}
		for _, s := range r.Samples {
			if !s.Valid {
				continue
			}
			dst := g.a
			if s.Arm == "B" {
				dst = g.b
			}
			for k, v := range s.Metrics {
				dst[k] = append(dst[k], v)
			}
		}
	}
	var b strings.Builder
	b.WriteString("## Resultados combinados\n\n")
	b.WriteString("Junta todas las ventanas válidas de todos los experimentos de cada cambio. Como aquí se miran los datos más de una vez, el umbral es más estricto (p < 0,01).\n\n")
	b.WriteString("| Cambio | Experimentos | Ventanas A/B | Métrica | A | B | Cambio | p | ¿Real? |\n|---|---|---|---|---|---|---|---|---|\n")
	for _, key := range order {
		g := groups[key]
		name := catalog.Title(g.tweak)
		if g.stress != "" {
			name += " (con CPU cargada)"
		}
		keys := make([]string, 0, len(g.primary))
		for k := range g.primary {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			a, bb := g.a[k], g.b[k]
			if len(a) == 0 || len(bb) == 0 {
				continue
			}
			c := stats.CompareAlpha(a, bb, 7, 0.01)
			mark := "no"
			if c.Conclusive {
				mark = "**sí**"
			}
			fmt.Fprintf(&b, "| %s | %d | %d/%d | %s | %s | %s | %s %% | %s | %s |\n", name, g.runs, len(a), len(bb), k, es(c.MedianA, 2), es(c.MedianB, 2), esSigned(c.DeltaPct, 1), es(c.PValue, 3), mark)
		}
	}
	b.WriteString("\n")
	return b.String()
}
