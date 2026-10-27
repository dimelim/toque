package frames

import (
	"math"
	"sort"

	"toque/internal/stats"
)

type Summary struct {
	Frames         int     `json:"frames"`
	Seconds        float64 `json:"seconds"`
	AvgFPS         float64 `json:"avgFps"`
	Low1FPS        float64 `json:"low1Fps"`
	Low01FPS       float64 `json:"low01Fps"`
	FrametimeP50   float64 `json:"frametimeP50"`
	FrametimeP99   float64 `json:"frametimeP99"`
	FrametimeStdev float64 `json:"frametimeStdev"`
	Stutters       int     `json:"stutters"`
	StutterPerMin  float64 `json:"stutterPerMin"`
	RenderLatency  float64 `json:"renderLatencyMs"`
	RenderLatP99   float64 `json:"renderLatencyP99Ms"`
	CPUBusy        float64 `json:"cpuBusyMs"`
	GPUBusy        float64 `json:"gpuBusyMs"`
	Bound          string  `json:"bound"`
	PresentMode    string  `json:"presentMode"`
	Tearing        bool    `json:"tearing"`
}

func valid(v float64) bool { return !math.IsNaN(v) && v > 0 }

func Summarize(fr []Frame) Summary {
	var s Summary
	var ft, lat, cpu, gpu []float64
	modes := map[string]int{}
	tear := 0
	for _, f := range fr {
		if !valid(f.BetweenPresents) {
			continue
		}
		ft = append(ft, f.BetweenPresents)
		if valid(f.CPUBusy) && valid(f.UntilDisplayed) {
			w := f.CPUWait
			if math.IsNaN(w) {
				w = 0
			}
			lat = append(lat, f.CPUBusy+w+f.UntilDisplayed)
		}
		if valid(f.CPUBusy) {
			cpu = append(cpu, f.CPUBusy)
		}
		if valid(f.GPUBusy) {
			gpu = append(gpu, f.GPUBusy)
		}
		modes[f.PresentMode]++
		if f.AllowsTearing {
			tear++
		}
	}
	if len(ft) == 0 {
		return s
	}
	total := 0.0
	for _, v := range ft {
		total += v
	}
	sorted := stats.Sorted(ft)
	s.Frames = len(ft)
	s.Seconds = total / 1000
	s.AvgFPS = float64(len(ft)) / s.Seconds
	s.Low1FPS = 1000 / stats.TopMean(ft, 0.01)
	s.Low01FPS = 1000 / stats.TopMean(ft, 0.001)
	s.FrametimeP50 = stats.Percentile(sorted, 50)
	s.FrametimeP99 = stats.Percentile(sorted, 99)
	s.FrametimeStdev = stats.Stdev(ft)
	s.Stutters = countStutters(ft)
	s.StutterPerMin = float64(s.Stutters) / (s.Seconds / 60)
	if len(lat) > 0 {
		s.RenderLatency = stats.Mean(lat)
		s.RenderLatP99 = stats.Percentile(stats.Sorted(lat), 99)
	}
	s.CPUBusy = stats.Mean(cpu)
	s.GPUBusy = stats.Mean(gpu)
	switch {
	case s.GPUBusy > 0.85*s.FrametimeP50:
		s.Bound = "gpu"
	case s.CPUBusy > 0.85*s.FrametimeP50:
		s.Bound = "cpu"
	default:
		s.Bound = "limiter"
	}
	type kv struct {
		k string
		v int
	}
	var list []kv
	for k, v := range modes {
		list = append(list, kv{k, v})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].v > list[j].v })
	if len(list) > 0 {
		s.PresentMode = list[0].k
	}
	s.Tearing = tear > len(ft)/2
	return s
}

func countStutters(ft []float64) int {
	const window = 30
	n := 0
	for i := range ft {
		lo := i - window
		if lo < 0 {
			lo = 0
		}
		hi := i + window
		if hi > len(ft) {
			hi = len(ft)
		}
		med := stats.Median(ft[lo:hi])
		if ft[i] > 2.5*med && ft[i] > med+8 {
			n++
		}
	}
	return n
}
