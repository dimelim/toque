package winx

import (
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	processPowerThrottling = 4
	throttleExecutionSpeed = 0x1
	throttleTimerIgnore    = 0x4
)

var (
	kernel32                  = windows.NewLazySystemDLL("kernel32.dll")
	procSetProcessInformation = kernel32.NewProc("SetProcessInformation")
	procGetProcessInformation = kernel32.NewProc("GetProcessInformation")
)

type powerThrottling struct {
	Version     uint32
	ControlMask uint32
	StateMask   uint32
}

type Eco struct {
	Class   uint32 `json:"class"`
	Control uint32 `json:"control"`
	State   uint32 `json:"state"`
	Known   bool   `json:"known"`
}

func EcoOf(pid uint32) (Eco, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return Eco{}, err
	}
	defer windows.CloseHandle(h)
	class, err := windows.GetPriorityClass(h)
	if err != nil {
		return Eco{}, err
	}
	e := Eco{Class: class}
	st := powerThrottling{Version: 1}
	if r, _, _ := procGetProcessInformation.Call(uintptr(h), processPowerThrottling, uintptr(unsafe.Pointer(&st)), unsafe.Sizeof(st)); r != 0 {
		e.Control, e.State, e.Known = st.ControlMask, st.StateMask, true
	}
	return e, nil
}

func SetEco(pid uint32, e Eco) error {
	h, err := windows.OpenProcess(windows.PROCESS_SET_INFORMATION|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	st := powerThrottling{Version: 1, ControlMask: e.Control, StateMask: e.State}
	if r, _, err := procSetProcessInformation.Call(uintptr(h), processPowerThrottling, uintptr(unsafe.Pointer(&st)), unsafe.Sizeof(st)); r == 0 {
		return err
	}
	return windows.SetPriorityClass(h, e.Class)
}

func Efficient() Eco {
	return Eco{Class: windows.IDLE_PRIORITY_CLASS, Control: throttleExecutionSpeed, State: throttleExecutionSpeed, Known: true}
}

func (e Eco) Restorable() Eco {
	if !e.Known {
		e.Control, e.State = 0, 0
	}
	switch e.Class {
	case windows.IDLE_PRIORITY_CLASS, windows.BELOW_NORMAL_PRIORITY_CLASS, windows.NORMAL_PRIORITY_CLASS, windows.ABOVE_NORMAL_PRIORITY_CLASS, windows.HIGH_PRIORITY_CLASS:
	default:
		e.Class = windows.NORMAL_PRIORITY_CLASS
	}
	e.Control &= throttleExecutionSpeed | throttleTimerIgnore
	e.State &= e.Control
	return e
}

func ProcTimes(pid uint32) (created int64, cpu time.Duration, err error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return 0, 0, err
	}
	defer windows.CloseHandle(h)
	var c, e, k, u windows.Filetime
	if err := windows.GetProcessTimes(h, &c, &e, &k, &u); err != nil {
		return 0, 0, err
	}
	return c.Nanoseconds(), time.Duration(k.Nanoseconds() + u.Nanoseconds()), nil
}

func SessionOf(pid uint32) (uint32, bool) {
	var sid uint32
	if err := windows.ProcessIdToSessionId(pid, &sid); err != nil {
		return 0, false
	}
	return sid, true
}
