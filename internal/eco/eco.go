package eco

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sort"
	"sync"

	"toque/internal/sysmon"
	"toque/internal/winx"
)

const minCore = 5

type Target struct {
	PID     uint32
	Created int64
	Exe     string
}

type entry struct {
	PID     uint32   `json:"pid"`
	Created int64    `json:"created"`
	Exe     string   `json:"exe"`
	Prev    winx.Eco `json:"prev"`
}

type Throttler struct {
	mu      sync.Mutex
	path    string
	entries map[uint32]entry
}

func New(path string) *Throttler {
	return &Throttler{path: path, entries: map[uint32]entry{}}
}

func Pick(procs []sysmon.ProcCPU, session uint32, audible map[uint32]bool, excused map[string]bool) []Target {
	use := map[string]float64{}
	loud := map[string]bool{}
	for _, p := range procs {
		if _, ok := apps[p.Name]; !ok || p.Session != session {
			continue
		}
		use[p.Name] += p.Core
		if audible[p.PID] {
			loud[p.Name] = true
		}
	}
	var out []Target
	for _, p := range procs {
		if _, ok := apps[p.Name]; !ok || p.Session != session {
			continue
		}
		if use[p.Name] < minCore || loud[p.Name] || excused[p.Name] {
			continue
		}
		out = append(out, Target{PID: p.PID, Created: p.Created, Exe: p.Name})
	}
	return out
}

func same(pid uint32, created int64) bool {
	c, _, err := winx.ProcTimes(pid)
	return err == nil && c == created
}

func (t *Throttler) Apply(targets []Target) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, tg := range targets {
		if _, ok := t.entries[tg.PID]; ok || !same(tg.PID, tg.Created) {
			continue
		}
		prev, err := winx.EcoOf(tg.PID)
		if err != nil {
			continue
		}
		t.entries[tg.PID] = entry{PID: tg.PID, Created: tg.Created, Exe: tg.Exe, Prev: prev}
		if t.save() != nil {
			delete(t.entries, tg.PID)
			continue
		}
		if winx.SetEco(tg.PID, winx.Efficient()) != nil {
			delete(t.entries, tg.PID)
			t.save()
		}
	}
}

func (t *Throttler) restore(keep func(entry) bool) error {
	var errs []error
	for pid, e := range t.entries {
		if keep(e) {
			continue
		}
		if same(e.PID, e.Created) {
			if err := winx.SetEco(e.PID, e.Prev.Restorable()); err != nil && winx.Alive(e.PID) {
				errs = append(errs, err)
				continue
			}
		}
		delete(t.entries, pid)
	}
	errs = append(errs, t.save())
	return errors.Join(errs...)
}

func (t *Throttler) RestoreAll() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.restore(func(entry) bool { return false })
}

func (t *Throttler) RestoreExe(exe string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.restore(func(e entry) bool { return e.Exe != exe })
}

func (t *Throttler) Recover() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	b, err := os.ReadFile(t.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var list []entry
	if err := json.Unmarshal(b, &list); err != nil {
		os.Remove(t.path)
		return err
	}
	for _, e := range list {
		t.entries[e.PID] = e
	}
	return t.restore(func(entry) bool { return false })
}

func (t *Throttler) Throttled() map[string]int {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := map[string]int{}
	for _, e := range t.entries {
		out[e.Exe]++
	}
	return out
}

func (t *Throttler) save() error {
	if len(t.entries) == 0 {
		err := os.Remove(t.path)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	list := make([]entry, 0, len(t.entries))
	for _, e := range t.entries {
		list = append(list, e)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].PID < list[j].PID })
	b, err := json.Marshal(list)
	if err != nil {
		return err
	}
	tmp := t.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, t.path)
}
