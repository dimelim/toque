package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"toque/internal/winx"
)

type dcEntry struct {
	RTT float64   `json:"rttMs"`
	At  time.Time `json:"at"`
	N   int       `json:"n"`
}

var dcMu sync.Mutex

func dcPath() string {
	return filepath.Join(winx.DataDir(), "datacenters.json")
}

func loadDatacenters() map[string]dcEntry {
	out := map[string]dcEntry{}
	if b, err := os.ReadFile(dcPath()); err == nil {
		json.Unmarshal(b, &out)
	}
	return out
}

func rememberDatacenter(ip string, rtt float64) {
	if ip == "" || rtt <= 0 {
		return
	}
	dcMu.Lock()
	defer dcMu.Unlock()
	all := loadDatacenters()
	e := all[ip]
	if e.N == 0 || time.Since(e.At) > 7*24*time.Hour {
		e = dcEntry{RTT: rtt, N: 1}
	} else {
		e.RTT = (e.RTT*float64(e.N) + rtt) / float64(e.N+1)
		e.N++
	}
	e.At = time.Now()
	all[ip] = e
	b, _ := json.MarshalIndent(all, "", "  ")
	os.WriteFile(dcPath(), b, 0o644)
}

type Datacenter struct {
	IP       string    `json:"ip"`
	RTT      float64   `json:"rttMs"`
	Measured int       `json:"measured"`
	At       time.Time `json:"at"`
}

func (s *Service) Datacenters() []Datacenter {
	dcMu.Lock()
	defer dcMu.Unlock()
	var out []Datacenter
	for ip, e := range loadDatacenters() {
		if time.Since(e.At) <= 7*24*time.Hour {
			out = append(out, Datacenter{IP: ip, RTT: e.RTT, Measured: e.N, At: e.At})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RTT < out[j].RTT })
	return out
}

func bestDatacenter() (float64, bool) {
	dcMu.Lock()
	defer dcMu.Unlock()
	best := 0.0
	for _, e := range loadDatacenters() {
		if time.Since(e.At) <= 7*24*time.Hour && (best == 0 || e.RTT < best) {
			best = e.RTT
		}
	}
	return best, best > 0
}

func knownDatacenter(ip string) (float64, bool) {
	dcMu.Lock()
	defer dcMu.Unlock()
	e, ok := loadDatacenters()[ip]
	if !ok || time.Since(e.At) > 7*24*time.Hour {
		return 0, false
	}
	return e.RTT, true
}
