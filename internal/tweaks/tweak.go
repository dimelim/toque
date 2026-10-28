package tweaks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"golang.org/x/sys/windows/registry"

	"toque/internal/i18n"
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
	Snapshot(Context) (Undo, error)
	Apply(Context) error
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

func ignoreMissing(err error) error {
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	return err
}

var journalMu sync.Mutex

type journal struct {
	Applied map[string]Undo `json:"applied"`
}

func journalPath() string {
	return filepath.Join(winx.DataDir(), "journal.json")
}

func loadJournal() (journal, error) {
	j := journal{Applied: map[string]Undo{}}
	b, err := os.ReadFile(journalPath())
	if errors.Is(err, fs.ErrNotExist) {
		return j, nil
	}
	if err != nil {
		return j, err
	}
	if err := json.Unmarshal(b, &j); err != nil {
		return journal{Applied: map[string]Undo{}}, fmt.Errorf(i18n.T("el diario de cambios %s está dañado (%w)"), journalPath(), err)
	}
	if j.Applied == nil {
		j.Applied = map[string]Undo{}
	}
	return j, nil
}

func (j journal) save() error {
	b, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return err
	}
	tmp := journalPath() + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, journalPath())
}

func (j journal) ids() []string {
	ids := make([]string, 0, len(j.Applied))
	for id := range j.Applied {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (j journal) revert(ctx Context, t Tweak) error {
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

func Entries() (map[string]Undo, error) {
	journalMu.Lock()
	defer journalMu.Unlock()
	j, err := loadJournal()
	return j.Applied, err
}

func Apply(ctx Context, t Tweak) error {
	journalMu.Lock()
	defer journalMu.Unlock()
	j, err := loadJournal()
	if err != nil {
		return err
	}
	if _, ok := j.Applied[t.ID()]; ok {
		return nil
	}
	undo, err := t.Snapshot(ctx)
	if err != nil {
		return err
	}
	j.Applied[t.ID()] = undo
	if err := j.save(); err != nil {
		return err
	}
	if err := t.Apply(ctx); err != nil {
		if j.revert(ctx, t) != nil {
			return fmt.Errorf(i18n.T("%w (el cambio quedó a medias y sigue en el diario para deshacerlo)"), err)
		}
		return err
	}
	return nil
}

func Revert(ctx Context, t Tweak) error {
	journalMu.Lock()
	defer journalMu.Unlock()
	j, err := loadJournal()
	if err != nil {
		return err
	}
	return j.revert(ctx, t)
}

func RevertAll(ctx Context) error {
	journalMu.Lock()
	defer journalMu.Unlock()
	j, err := loadJournal()
	if err != nil {
		return err
	}
	var errs []error
	for _, id := range j.ids() {
		t, err := Get(id)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if err := j.revert(ctx, t); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", id, err))
		}
	}
	return errors.Join(errs...)
}

func Resync(ctx Context) error {
	journalMu.Lock()
	defer journalMu.Unlock()
	j, err := loadJournal()
	if err != nil {
		return err
	}
	var errs []error
	forgot := false
	cur, flagsErr := roblox.ReadFlags(ctx.Install.VersionDir)
	if ctx.Install.VersionDir == "" {
		flagsErr = errors.New("roblox not installed")
	}
	missing := map[string]any{}
	for _, id := range j.ids() {
		t, err := Get(id)
		if err != nil {
			continue
		}
		switch f := t.(type) {
		case flagSet:
			for k, v := range f.flags {
				if flagsErr == nil && fmt.Sprint(cur[k]) != fmt.Sprint(v) {
					missing[k] = v
				}
			}
		case basicSetting:
			if !t.State(ctx).Applied {
				delete(j.Applied, id)
				forgot = true
			}
		}
	}
	if len(missing) > 0 {
		if len(roblox.Running()) > 0 {
			errs = append(errs, errRobloxOpen)
		} else {
			maps.Copy(cur, missing)
			errs = append(errs, roblox.WriteFlags(ctx.Install.VersionDir, cur))
		}
	}
	if forgot {
		errs = append(errs, j.save())
	}
	return errors.Join(errs...)
}
