package bench

import (
	"sort"

	"toque/internal/stats"
)

const PooledAlpha = 0.01

type Pool struct {
	Tweak   string
	Stress  string
	Results []ABResult
	A, B    map[string][]float64
	primary map[string]bool
}

func PoolResults(results []ABResult) []*Pool {
	byKey := map[string]*Pool{}
	var out []*Pool
	for _, r := range results {
		key := r.Config.Tweak + "|" + r.Config.Stress
		p, ok := byKey[key]
		if !ok {
			p = &Pool{Tweak: r.Config.Tweak, Stress: r.Config.Stress, A: map[string][]float64{}, B: map[string][]float64{}, primary: map[string]bool{}}
			byKey[key] = p
			out = append(out, p)
		}
		p.Results = append(p.Results, r)
		for _, k := range r.Config.Primary {
			p.primary[k] = true
		}
		for _, s := range r.Samples {
			if !s.Valid {
				continue
			}
			dst := p.A
			if s.Arm == "B" {
				dst = p.B
			}
			for k, v := range s.Metrics {
				dst[k] = append(dst[k], v)
			}
		}
	}
	return out
}

func (p *Pool) Primary() []string {
	keys := make([]string, 0, len(p.primary))
	for k := range p.primary {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (p *Pool) Compare(metric string) (stats.Comparison, bool) {
	a, b := p.A[metric], p.B[metric]
	if len(a) == 0 || len(b) == 0 {
		return stats.Comparison{}, false
	}
	return stats.CompareAlpha(a, b, 7, PooledAlpha), true
}

func (p *Pool) Verdict() string {
	improved, worsened := false, false
	for _, k := range p.Primary() {
		c, ok := p.Compare(k)
		if !ok || !c.Conclusive {
			continue
		}
		if c.DeltaPct*float64(Better[k]) > 0 {
			improved = true
		} else {
			worsened = true
		}
	}
	switch {
	case worsened:
		return "worse"
	case improved:
		return "improves"
	}
	return "no-effect"
}
