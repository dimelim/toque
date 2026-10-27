package sysmon

import (
	"sort"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type ProcMem struct {
	Name string  `json:"name"`
	MB   float64 `json:"mb"`
	PIDs int     `json:"count"`
}

func privateMB(pid uint32) float64 {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return 0
	}
	defer windows.CloseHandle(h)
	var mc memCounters
	mc.Cb = uint32(unsafe.Sizeof(mc))
	if r, _, _ := procGetProcessMemInfo.Call(uintptr(h), uintptr(unsafe.Pointer(&mc)), uintptr(mc.Cb)); r == 0 {
		return 0
	}
	return float64(mc.WorkingSetSize) / (1 << 20)
}

var systemProcs = map[string]bool{
	"system": true, "registry": true, "memory compression": true, "idle": true, "secure system": true,
	"csrss.exe": true, "smss.exe": true, "wininit.exe": true, "services.exe": true, "lsass.exe": true,
	"dwm.exe": true, "svchost.exe": true, "msmpeng.exe": true, "fontdrvhost.exe": true, "winlogon.exe": true,
}

type Proc struct {
	PID  uint32
	Name string
}

func TopMemory(list []Proc, n int, skip func(string) bool) []ProcMem {
	agg := map[string]*ProcMem{}
	for _, p := range list {
		key := strings.ToLower(p.Name)
		if systemProcs[key] || (skip != nil && skip(key)) {
			continue
		}
		mb := privateMB(p.PID)
		if mb <= 0 {
			continue
		}
		pm, ok := agg[key]
		if !ok {
			pm = &ProcMem{Name: strings.TrimSuffix(p.Name, ".exe")}
			agg[key] = pm
		}
		pm.MB += mb
		pm.PIDs++
	}
	out := make([]ProcMem, 0, len(agg))
	for _, v := range agg {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].MB > out[j].MB })
	if len(out) > n {
		out = out[:n]
	}
	return out
}
