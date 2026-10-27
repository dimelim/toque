package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"golang.org/x/sys/windows"

	"toque/internal/bench"
	"toque/internal/catalog"
	"toque/internal/winx"
)

type RunRecord struct {
	ID        string       `json:"id"`
	Label     string       `json:"label"`
	At        time.Time    `json:"at"`
	ServerIP  string       `json:"serverIp"`
	JobID     string       `json:"jobId"`
	JoinMs    int64        `json:"joinMs"`
	Applied   []string     `json:"applied"`
	Window    bench.Window `json:"window"`
}

func runsDir() string {
	return winx.DataDir("history")
}

func SaveRun(r RunRecord) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(runsDir(), r.ID+".json"), b, 0o644)
}

func LoadRuns() []RunRecord {
	files, _ := filepath.Glob(filepath.Join(runsDir(), "*.json"))
	var out []RunRecord
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var r RunRecord
		if json.Unmarshal(b, &r) == nil {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	return out
}

func DeleteRun(id string) error {
	return os.Remove(filepath.Join(runsDir(), filepath.Base(id)+".json"))
}

type ExperimentRecord struct {
	ID      string           `json:"id"`
	Tweak   string           `json:"tweak"`
	Title   string           `json:"title"`
	At      time.Time        `json:"at"`
	Verdict string           `json:"verdict"`
	Summary string           `json:"summary"`
	Valid   int              `json:"valid"`
	Stress  string           `json:"stress"`
	Primary []string         `json:"primary"`
	Compare map[string]Delta `json:"compare"`
}

var planTitles = map[string]string{
	"power-plan:381b4222-f694-41f0-9685-ff5bb260df2e": "Plan de energía Equilibrado",
	"power-plan:8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c": "Plan de energía Alto rendimiento",
	"power-plan:e9a42b02-d5df-448d-aa00-03f14749eb61": "Plan de energía Máximo rendimiento",
	"proc-priority-high":                              "Prioridad alta para Roblox (en vivo)",
}

func experimentTitle(id string) string {
	if e, ok := catalog.Find(id); ok {
		return e.Title
	}
	if t, ok := planTitles[id]; ok {
		return t
	}
	return id
}

type Delta struct {
	A          float64 `json:"a"`
	B          float64 `json:"b"`
	Pct        float64 `json:"pct"`
	Low        float64 `json:"low"`
	High       float64 `json:"high"`
	P          float64 `json:"p"`
	Conclusive bool    `json:"conclusive"`
}

func toRecord(res bench.ABResult) ExperimentRecord {
	rec := ExperimentRecord{
		ID: filepath.Base(res.Dir), Tweak: res.Config.Tweak, Title: experimentTitle(res.Config.Tweak), At: res.Started,
		Verdict: res.Verdict, Summary: res.Summary, Stress: res.Config.Stress, Primary: res.Config.Primary, Compare: map[string]Delta{},
	}
	for _, smp := range res.Samples {
		if smp.Valid {
			rec.Valid++
		}
	}
	for k, c := range res.Compare {
		rec.Compare[k] = Delta{A: c.MedianA, B: c.MedianB, Pct: c.DeltaPct, Low: c.CILowPct, High: c.CIHighPct, P: c.PValue, Conclusive: c.Conclusive}
	}
	return rec
}

func LoadExperiments() []ExperimentRecord {
	dirs, _ := filepath.Glob(filepath.Join(winx.DataDir("experiments"), "*", "result.json"))
	var out []ExperimentRecord
	for _, f := range dirs {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var res bench.ABResult
		if json.Unmarshal(b, &res) == nil {
			res.Compare, res.Verdict, res.Summary = bench.Evaluate(res.Samples, res.Config.Primary)
			out = append(out, toRecord(res))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	return out
}

var procGetTickCount64 = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetTickCount64")

func uptime() time.Duration {
	ms, _, _ := procGetTickCount64.Call()
	return time.Duration(ms) * time.Millisecond
}
