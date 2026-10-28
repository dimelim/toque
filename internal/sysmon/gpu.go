package sysmon

import (
	"strconv"
	"strings"
)

type GPUSampler struct {
	q *query
	c uintptr
}

func NewGPUSampler() (*GPUSampler, error) {
	q, err := openQuery()
	if err != nil {
		return nil, err
	}
	c, err := q.add(`\GPU Engine(*)\Utilization Percentage`)
	if err != nil {
		q.close()
		return nil, err
	}
	q.collect()
	return &GPUSampler{q: q, c: c}, nil
}

func (g *GPUSampler) Sample() map[uint32]float64 {
	g.q.collect()
	return gpuByProcess(g.q.array(g.c))
}

func (g *GPUSampler) Close() {
	g.q.close()
}

func gpuByProcess(engines map[string]float64) map[uint32]float64 {
	byType := map[uint32]map[string]float64{}
	for name, v := range engines {
		pid, engine, ok := parseEngine(name)
		if !ok {
			continue
		}
		if byType[pid] == nil {
			byType[pid] = map[string]float64{}
		}
		byType[pid][engine] += v
	}
	out := make(map[uint32]float64, len(byType))
	for pid, types := range byType {
		for _, v := range types {
			out[pid] = max(out[pid], v)
		}
	}
	return out
}

func parseEngine(name string) (uint32, string, bool) {
	rest, ok := strings.CutPrefix(name, "pid_")
	if !ok {
		return 0, "", false
	}
	pidText, rest, ok := strings.Cut(rest, "_")
	if !ok {
		return 0, "", false
	}
	pid, err := strconv.ParseUint(pidText, 10, 32)
	_, engine, found := strings.Cut(rest, "engtype_")
	if err != nil || !found {
		return 0, "", false
	}
	return uint32(pid), engine, true
}
