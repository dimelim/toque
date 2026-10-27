package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"toque/internal/bench"
	"toque/internal/catalog"
)

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
	var results []bench.ABResult
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var r bench.ABResult
		if json.Unmarshal(raw, &r) == nil {
			results = append(results, r)
		}
	}
	var b strings.Builder
	b.WriteString("## Resultados combinados\n\n")
	b.WriteString("Junta todas las ventanas válidas de todos los experimentos de cada cambio. Como aquí se miran los datos más de una vez, el umbral es más estricto (p < 0,01).\n\n")
	b.WriteString("| Cambio | Experimentos | Ventanas A/B | Métrica | A | B | Cambio | p | ¿Real? |\n|---|---|---|---|---|---|---|---|---|\n")
	for _, p := range bench.PoolResults(results) {
		name := catalog.Title(p.Tweak)
		if p.Stress != "" {
			name += " (con CPU cargada)"
		}
		for _, k := range p.Primary() {
			c, ok := p.Compare(k)
			if !ok {
				continue
			}
			mark := "no"
			if c.Conclusive {
				mark = "**sí**"
			}
			fmt.Fprintf(&b, "| %s | %d | %d/%d | %s | %s | %s | %s %% | %s | %s |\n", name, len(p.Results), len(p.A[k]), len(p.B[k]), k, es(c.MedianA, 2), es(c.MedianB, 2), esSigned(c.DeltaPct, 1), es(c.PValue, 3), mark)
		}
	}
	b.WriteString("\n")
	return b.String()
}
