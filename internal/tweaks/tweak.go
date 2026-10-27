package tweaks

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"toque/internal/roblox"
	"toque/internal/winx"
)

type Kind string

const (
	Live     Kind = "live"
	Relaunch Kind = "relaunch"
	Reboot   Kind = "reboot"
)

type Context struct {
	Install roblox.Install
	PID     uint32
}

type State struct {
	Applicable bool   `json:"applicable"`
	Applied    bool   `json:"applied"`
	Detail     string `json:"detail"`
}

type Undo map[string]string

type Tweak interface {
	ID() string
	Kind() Kind
	State(Context) State
	Apply(Context) (Undo, error)
	Revert(Context, Undo) error
}

var (
	registryMu sync.Mutex
	registered = map[string]Tweak{}
)

func Register(t Tweak) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registered[t.ID()] = t
}

func Get(id string) (Tweak, error) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if t, ok := registered[id]; ok {
		return t, nil
	}
	if t, err := parametric(id); err == nil {
		return t, nil
	}
	return nil, fmt.Errorf("unknown tweak %q", id)
}

func IDs() []string {
	registryMu.Lock()
	defer registryMu.Unlock()
	var out []string
	for id := range registered {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

type Journal struct {
	path    string
	Applied map[string]Undo `json:"applied"`
}

func OpenJournal() *Journal {
	j := &Journal{path: filepath.Join(winx.DataDir(), "journal.json"), Applied: map[string]Undo{}}
	if b, err := os.ReadFile(j.path); err == nil {
		json.Unmarshal(b, j)
		if j.Applied == nil {
			j.Applied = map[string]Undo{}
		}
	}
	return j
}

func (j *Journal) save() error {
	b, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return err
	}
	tmp := j.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, j.path)
}

func (j *Journal) Apply(ctx Context, t Tweak) error {
	if _, ok := j.Applied[t.ID()]; ok {
		return nil
	}
	undo, err := t.Apply(ctx)
	if err != nil {
		return err
	}
	j.Applied[t.ID()] = undo
	return j.save()
}

func (j *Journal) Revert(ctx Context, t Tweak) error {
	undo, ok := j.Applied[t.ID()]
	if !ok {
		return nil
	}
	if err := t.Revert(ctx, undo); err != nil {
		return err
	}
	delete(j.Applied, t.ID())
	return j.save()
}

func (j *Journal) RevertAll(ctx Context) error {
	var errs []error
	for id := range j.Applied {
		t, err := Get(id)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if err := j.Revert(ctx, t); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", id, err))
		}
	}
	return errors.Join(errs...)
}

func (j *Journal) Has(id string) bool {
	_, ok := j.Applied[id]
	return ok
}
