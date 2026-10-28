package app

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sys/windows"

	"toque/internal/tweaks"
	"toque/internal/winx"
)

const (
	visualID     = "session:windows-fx"
	powerID      = "session:power-overlay"
	balancedPlan = "381b4222-f694-41f0-9685-ff5bb260df2e"
)

type systemChange[T any] struct {
	mu     sync.Mutex
	file   string
	read   func() (T, error)
	target func(T) (T, bool)
	write  func(T) error
}

var systemDir = func() string { return winx.DataDir() }

func (c *systemChange[T]) path() string { return filepath.Join(systemDir(), c.file) }

func (c *systemChange[T]) applied() bool {
	_, err := os.Stat(c.path())
	return err == nil
}

func (c *systemChange[T]) on() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.applied() {
		return nil
	}
	cur, err := c.read()
	if err != nil {
		return err
	}
	next, needed := c.target(cur)
	if !needed {
		return nil
	}
	b, err := json.Marshal(cur)
	if err != nil {
		return err
	}
	if err := os.WriteFile(c.path(), b, 0o644); err != nil {
		return err
	}
	return c.write(next)
}

func (c *systemChange[T]) off() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	b, err := os.ReadFile(c.path())
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var prev T
	if err := json.Unmarshal(b, &prev); err != nil {
		os.Remove(c.path())
		return err
	}
	if err := c.write(prev); err != nil {
		return err
	}
	return os.Remove(c.path())
}

var visualChange = &systemChange[winx.VisualFX]{
	file:   "visual.json",
	read:   winx.ReadVisualFX,
	target: func(cur winx.VisualFX) (winx.VisualFX, bool) { return cur.Off(), !cur.AllOff() },
	write:  winx.ApplyVisualFX,
}

var powerChange = &systemChange[windows.GUID]{
	file: "power.json",
	read: winx.PowerOverlay,
	target: func(cur windows.GUID) (windows.GUID, bool) {
		return winx.BestPerformance, cur != winx.BestPerformance && powerApplicable()
	},
	write: winx.SetPowerOverlay,
}

func powerApplicable() bool {
	laptop, plugged := winx.Laptop()
	plan, err := tweaks.ActivePlan()
	return laptop && plugged && err == nil && strings.EqualFold(plan, balancedPlan)
}

type switcher interface {
	on() error
	off() error
	applied() bool
}

type switchTweak struct {
	id string
	sw switcher
}

func (t switchTweak) ID() string      { return t.id }
func (switchTweak) Kind() tweaks.Kind { return tweaks.Live }
func (switchTweak) Snapshot(tweaks.Context) (tweaks.Undo, error) {
	return tweaks.Undo{}, nil
}

func (t switchTweak) State(tweaks.Context) tweaks.State {
	return tweaks.State{Applicable: winx.IsElevated(), Applied: t.sw.applied()}
}

func (t switchTweak) Apply(tweaks.Context) error {
	return t.sw.on()
}

func (t switchTweak) Revert(tweaks.Context, tweaks.Undo) error {
	return t.sw.off()
}

var systemTweaks = []switchTweak{{visualID, visualChange}, {powerID, powerChange}}

func restoreSystem() error {
	var errs []error
	for _, t := range systemTweaks {
		errs = append(errs, t.sw.off())
	}
	return errors.Join(errs...)
}

func init() {
	for _, t := range systemTweaks {
		tweaks.Register(t)
	}
}
