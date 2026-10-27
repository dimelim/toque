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
	Mbps   float64 `json:"mbps"`
}

type BloatResult struct {
	Target  string      `json:"target"`
	Phases  []LoadPhase `json:"phases"`
	AddedMs float64     `json:"addedMs"`
	MaxLoss float64     `json:"maxLossPct"`
	Grade   string      `json:"grade"`
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

var grades = []string{"A+", "A", "B", "C", "D", "F"}

func latencyGrade(added float64) int {
	switch {
	case added < 5:
		return 0
	case added < 30:
		return 1
	case added < 60:
		return 2
	case added < 200:
		return 3
	case added < 400:
		return 4
	}
	return 5
}

func lossGrade(loss float64) int {
	switch {
	case loss < 2:
		return 0
	case loss < 5:
		return 2
	case loss < 10:
		return 3
	case loss < 20:
		return 4
	}
	return 5
}

func Bufferbloat(ctx context.Context, dst netip.Addr, phase time.Duration, load func(kind string) func() int64) BloatResult {
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
		start := time.Now()
		p := sample(ctx, dst, phase)
		bytes := stop()
		if secs := time.Since(start).Seconds() + 2; secs > 0 {
			p.Mbps = float64(bytes) * 8 / 1e6 / secs
		}
		if kind == "down" {
			p.Phase = "descarga"
		} else {
			p.Phase = "subida"
		}
		res.Phases = append(res.Phases, p)
		if p.P95 > worst {
			worst = p.P95
		}
		if p.Loss > res.MaxLoss {
			res.MaxLoss = p.Loss
		}
		time.Sleep(time.Second)
	}
	res.AddedMs = worst - idle.RTT
	if res.AddedMs < 0 {
		res.AddedMs = 0
	}
	res.Grade = grades[max(latencyGrade(res.AddedMs), lossGrade(res.MaxLoss))]
	return res
}
