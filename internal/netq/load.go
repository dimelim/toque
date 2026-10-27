package netq

import (
	"context"
	"net/netip"
	"sort"
	"time"

	"toque/internal/stats"
)

type LoadPhase struct {
	Phase  string  `json:"phase"`
	RTT    float64 `json:"rttMs"`
	P95    float64 `json:"p95Ms"`
	Jitter float64 `json:"jitterMs"`
	Loss   float64 `json:"lossPct"`
}

type BloatResult struct {
	Target   string      `json:"target"`
	Phases   []LoadPhase `json:"phases"`
	AddedMs  float64     `json:"addedMs"`
	Grade    string      `json:"grade"`
}

func sample(ctx context.Context, dst netip.Addr, dur time.Duration) LoadPhase {
	var rtts []float64
	sent, lost := 0, 0
	deadline := time.Now().Add(dur)
	for time.Now().Before(deadline) && ctx.Err() == nil {
		sent++
		e, err := echo(dst, 128, 800*time.Millisecond)
		if err != nil || !e.Reached {
			lost++
		} else {
			rtts = append(rtts, float64(e.RTT.Microseconds())/1000)
		}
		time.Sleep(120 * time.Millisecond)
	}
	p := LoadPhase{}
	if sent > 0 {
		p.Loss = 100 * float64(lost) / float64(sent)
	}
	if len(rtts) > 0 {
		s := append([]float64(nil), rtts...)
		sort.Float64s(s)
		p.RTT = stats.Percentile(s, 50)
		p.P95 = stats.Percentile(s, 95)
		p.Jitter = stats.Stdev(rtts)
	}
	return p
}

func grade(added float64) string {
	switch {
	case added < 5:
		return "A+"
	case added < 30:
		return "A"
	case added < 60:
		return "B"
	case added < 200:
		return "C"
	case added < 400:
		return "D"
	}
	return "F"
}

func Bufferbloat(ctx context.Context, dst netip.Addr, phase time.Duration, load func(kind string) func()) BloatResult {
	res := BloatResult{Target: dst.String()}
	idle := sample(ctx, dst, phase)
	idle.Phase = "reposo"
	res.Phases = append(res.Phases, idle)
	worst := idle.P95
	for _, kind := range []string{"down", "up"} {
		if ctx.Err() != nil {
			break
		}
		stop := load(kind)
		time.Sleep(2 * time.Second)
		p := sample(ctx, dst, phase)
		stop()
		if kind == "down" {
			p.Phase = "descarga"
		} else {
			p.Phase = "subida"
		}
		res.Phases = append(res.Phases, p)
		if p.P95 > worst {
			worst = p.P95
		}
		time.Sleep(time.Second)
	}
	res.AddedMs = worst - idle.RTT
	if res.AddedMs < 0 {
		res.AddedMs = 0
	}
	res.Grade = grade(res.AddedMs)
	return res
}
