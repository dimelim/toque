package sysmon

import (
	"strings"
	"time"

	"toque/internal/winx"
)

type ProcCPU struct {
	PID     uint32
	Parent  uint32
	Name    string
	Created int64
	Core    float64
	RAMMB   float64
	Session uint32
}

type CPUSampler struct {
	last map[uint32]cpuMark
	at   time.Time
}

type cpuMark struct {
	created int64
	cpu     time.Duration
}

func (s *CPUSampler) Reset() {
	s.last, s.at = nil, time.Time{}
}

func (s *CPUSampler) Sample() []ProcCPU {
	procs, err := winx.Processes()
	if err != nil {
		return nil
	}
	now := time.Now()
	elapsed := now.Sub(s.at)
	next := make(map[uint32]cpuMark, len(procs))
	out := make([]ProcCPU, 0, len(procs))
	for _, p := range procs {
		if p.PID == 0 || p.PID == 4 {
			continue
		}
		created, cpu, err := winx.ProcTimes(p.PID)
		if err != nil {
			continue
		}
		next[p.PID] = cpuMark{created: created, cpu: cpu}
		pc := ProcCPU{PID: p.PID, Parent: p.ParentPID, Name: strings.ToLower(p.Name), Created: created, RAMMB: privateMB(p.PID)}
		pc.Session, _ = winx.SessionOf(p.PID)
		if prev, ok := s.last[p.PID]; ok && prev.created == created && elapsed > 0 {
			pc.Core = 100 * float64(cpu-prev.cpu) / float64(elapsed)
		}
		out = append(out, pc)
	}
	s.last, s.at = next, now
	return out
}
