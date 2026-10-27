package tweaks

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	powrprof                 = windows.NewLazySystemDLL("powrprof.dll")
	procPowerGetActiveScheme = powrprof.NewProc("PowerGetActiveScheme")
	procPowerSetActiveScheme = powrprof.NewProc("PowerSetActiveScheme")
	ntdll                    = windows.NewLazySystemDLL("ntdll.dll")
	procNtSetTimerResolution = ntdll.NewProc("NtSetTimerResolution")
	procNtQueryTimerRes      = ntdll.NewProc("NtQueryTimerResolution")
)

const (
	PlanBalanced = "381b4222-f694-41f0-9685-ff5bb260df2e"
	PlanHigh     = "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c"
)

func ActivePlan() (string, error) {
	var g *windows.GUID
	if r, _, _ := procPowerGetActiveScheme.Call(0, uintptr(unsafe.Pointer(&g))); r != 0 || g == nil {
		return "", errors.New("PowerGetActiveScheme failed")
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(g)))
	return strings.ToLower(strings.Trim(g.String(), "{}")), nil
}

func SetPlan(id string) error {
	g, err := windows.GUIDFromString("{" + id + "}")
	if err != nil {
		return err
	}
	if r, _, _ := procPowerSetActiveScheme.Call(0, uintptr(unsafe.Pointer(&g))); r != 0 {
		return errors.New("PowerSetActiveScheme failed: " + strconv.Itoa(int(r)))
	}
	return nil
}

type powerPlan struct {
	target string
}

func (p powerPlan) ID() string { return "power-plan:" + p.target }
func (powerPlan) Kind() Kind   { return Live }

func (p powerPlan) State(Context) State {
	cur, err := ActivePlan()
	if err != nil {
		return State{Detail: err.Error()}
	}
	return State{Applicable: true, Applied: cur == p.target, Detail: "active " + cur}
}

func (p powerPlan) Apply(Context) (Undo, error) {
	cur, err := ActivePlan()
	if err != nil {
		return nil, err
	}
	return Undo{"prev": cur}, SetPlan(p.target)
}

func (p powerPlan) Revert(_ Context, u Undo) error {
	return SetPlan(u["prev"])
}

var timerMu sync.Mutex

func TimerResolution() (min, max, cur uint32) {
	procNtQueryTimerRes.Call(uintptr(unsafe.Pointer(&min)), uintptr(unsafe.Pointer(&max)), uintptr(unsafe.Pointer(&cur)))
	return
}

type timerRes struct{}

func (timerRes) ID() string { return "timer-res-0.5ms" }
func (timerRes) Kind() Kind { return Live }

func (timerRes) State(Context) State {
	_, _, cur := TimerResolution()
	return State{Applicable: true, Applied: cur <= 5000, Detail: "current " + strconv.FormatFloat(float64(cur)/10000, 'f', 3, 64) + " ms"}
}

func (timerRes) Apply(Context) (Undo, error) {
	timerMu.Lock()
	defer timerMu.Unlock()
	var actual uint32
	if r, _, _ := procNtSetTimerResolution.Call(5000, 1, uintptr(unsafe.Pointer(&actual))); r != 0 {
		return nil, errors.New("NtSetTimerResolution failed")
	}
	return Undo{"actual": strconv.FormatUint(uint64(actual), 10)}, nil
}

func (timerRes) Revert(Context, Undo) error {
	timerMu.Lock()
	defer timerMu.Unlock()
	var actual uint32
	procNtSetTimerResolution.Call(5000, 0, uintptr(unsafe.Pointer(&actual)))
	return nil
}

func init() {
	Register(timerRes{})
}
