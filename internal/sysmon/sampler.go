package sysmon

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"toque/internal/stats"
)

var (
	kernel32              = windows.NewLazySystemDLL("kernel32.dll")
	psapi                 = windows.NewLazySystemDLL("psapi.dll")
	procGetSystemTimes    = kernel32.NewProc("GetSystemTimes")
	procGetProcessMemInfo = psapi.NewProc("GetProcessMemoryInfo")
)

type Sample struct {
	CPUTotal   float64
	CPUProc    float64
	GPUTotal   float64
	GPUProc    float64
	AvailMB    float64
	ProcMB     float64
	HardFaults float64
	DiskReadMs map[string]float64
	DiskReads  map[string]float64
}

type Summary struct {
	Samples       int     `json:"samples"`
	CPUTotal      float64 `json:"cpuTotalPct"`
	CPUTotalMax   float64 `json:"cpuTotalMaxPct"`
	CPUProc       float64 `json:"cpuRobloxPct"`
	GPUTotal      float64 `json:"gpuTotalPct"`
	GPUProc       float64 `json:"gpuRobloxPct"`
	AvailMB       float64 `json:"ramAvailMb"`
	AvailMinMB    float64 `json:"ramAvailMinMb"`
	ProcMB        float64 `json:"ramRobloxMb"`
	HardFaults    float64 `json:"hardFaultsPerSec"`
	HardFaultsMax float64 `json:"hardFaultsMaxPerSec"`
	Disks         []Disk  `json:"disks"`
}

type Disk struct {
	Name      string  `json:"name"`
	ReadsSec  float64 `json:"readsPerSec"`
	ReadMs    float64 `json:"readLatencyMs"`
	ReadMsMax float64 `json:"readLatencyMaxMs"`
}

type memCounters struct {
	Cb                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
	PrivateUsage               uintptr
}

func ftU64(f windows.Filetime) uint64 {
	return uint64(f.HighDateTime)<<32 | uint64(f.LowDateTime)
}

func systemTimes() (idle, total uint64) {
	var i, k, u windows.Filetime
	procGetSystemTimes.Call(uintptr(unsafe.Pointer(&i)), uintptr(unsafe.Pointer(&k)), uintptr(unsafe.Pointer(&u)))
	return ftU64(i), ftU64(k) + ftU64(u)
}

func processTimes(h windows.Handle) uint64 {
	var c, e, k, u windows.Filetime
	if windows.GetProcessTimes(h, &c, &e, &k, &u) != nil {
		return 0
	}
	return ftU64(k) + ftU64(u)
}

func Run(ctx context.Context, pid uint32, interval time.Duration) ([]Sample, error) {
	q, err := openQuery()
	if err != nil {
		return nil, err
	}
	defer q.close()
	gpu, _ := q.add(`\GPU Engine(*)\Utilization Percentage`)
	faults, _ := q.add(`\Memory\Page Reads/sec`)
	diskLat, _ := q.add(`\PhysicalDisk(*)\Avg. Disk sec/Read`)
	diskReads, _ := q.add(`\PhysicalDisk(*)\Disk Reads/sec`)
	var proc windows.Handle
	if pid != 0 {
		proc, _ = windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if proc != 0 {
			defer windows.CloseHandle(proc)
		}
	}
	pidTag := fmt.Sprintf("pid_%d_", pid)
	cores := float64(runtime.NumCPU())
	q.collect()
	prevIdle, prevTotal := systemTimes()
	prevProc := processTimes(proc)
	prevWall := time.Now()
	var out []Sample
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return out, nil
		case <-tick.C:
		}
		q.collect()
		idle, total := systemTimes()
		now := time.Now()
		var s Sample
		if dt := float64(total - prevTotal); dt > 0 {
			s.CPUTotal = 100 * (1 - float64(idle-prevIdle)/dt)
		}
		if proc != 0 {
			cur := processTimes(proc)
			wall := now.Sub(prevWall).Seconds() * 1e7
			if wall > 0 {
				s.CPUProc = 100 * float64(cur-prevProc) / wall / cores
			}
			prevProc = cur
			var mc memCounters
			mc.Cb = uint32(unsafe.Sizeof(mc))
			if r, _, _ := procGetProcessMemInfo.Call(uintptr(proc), uintptr(unsafe.Pointer(&mc)), uintptr(mc.Cb)); r != 0 {
				s.ProcMB = float64(mc.PrivateUsage) / (1 << 20)
			}
		}
		prevIdle, prevTotal, prevWall = idle, total, now
		perEngine := map[string]float64{}
		for name, v := range q.array(gpu) {
			if !strings.Contains(name, "engtype_3D") {
				continue
			}
			key := name
			if j := strings.Index(name, "luid_"); j > 0 {
				key = name[j:]
			}
			perEngine[key] += v
			if strings.HasPrefix(name, pidTag) {
				s.GPUProc += v
			}
		}
		for _, v := range perEngine {
			if v > s.GPUTotal {
				s.GPUTotal = v
			}
		}
		if s.GPUTotal > 100 {
			s.GPUTotal = 100
		}
		if s.GPUProc > 100 {
			s.GPUProc = 100
		}
		if v, ok := q.value(faults); ok {
			s.HardFaults = v
		}
		s.AvailMB = AvailableMB()
		s.DiskReadMs = map[string]float64{}
		s.DiskReads = map[string]float64{}
		for name, v := range q.array(diskLat) {
			if name != "_Total" {
				s.DiskReadMs[name] = v * 1000
			}
		}
		for name, v := range q.array(diskReads) {
			if name != "_Total" {
				s.DiskReads[name] = v
			}
		}
		out = append(out, s)
	}
}

func Summarize(samples []Sample) Summary {
	var s Summary
	s.Samples = len(samples)
	if len(samples) == 0 {
		return s
	}
	col := func(f func(Sample) float64) []float64 {
		v := make([]float64, len(samples))
		for i, x := range samples {
			v[i] = f(x)
		}
		return v
	}
	cpu := col(func(x Sample) float64 { return x.CPUTotal })
	avail := col(func(x Sample) float64 { return x.AvailMB })
	hf := col(func(x Sample) float64 { return x.HardFaults })
	s.CPUTotal = stats.Mean(cpu)
	s.CPUTotalMax = stats.Sorted(cpu)[len(cpu)-1]
	s.CPUProc = stats.Mean(col(func(x Sample) float64 { return x.CPUProc }))
	s.GPUTotal = stats.Mean(col(func(x Sample) float64 { return x.GPUTotal }))
	s.GPUProc = stats.Mean(col(func(x Sample) float64 { return x.GPUProc }))
	s.AvailMB = stats.Mean(avail)
	s.AvailMinMB = stats.Sorted(avail)[0]
	s.ProcMB = stats.Mean(col(func(x Sample) float64 { return x.ProcMB }))
	s.HardFaults = stats.Mean(hf)
	s.HardFaultsMax = stats.Sorted(hf)[len(hf)-1]
	names := map[string]bool{}
	for _, x := range samples {
		for n := range x.DiskReads {
			names[n] = true
		}
	}
	for n := range names {
		var reads, lat []float64
		weighted, totalReads := 0.0, 0.0
		for _, x := range samples {
			r := x.DiskReads[n]
			reads = append(reads, r)
			if r > 0 {
				lat = append(lat, x.DiskReadMs[n])
				weighted += x.DiskReadMs[n] * r
				totalReads += r
			}
		}
		d := Disk{Name: n, ReadsSec: stats.Mean(reads)}
		if totalReads > 0 {
			d.ReadMs = weighted / totalReads
			d.ReadMsMax = stats.Sorted(lat)[len(lat)-1]
		}
		s.Disks = append(s.Disks, d)
	}
	return s
}
