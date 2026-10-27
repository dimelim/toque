package tweaks

import (
	"errors"
	"strconv"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type procPriority struct {
	class uint32
}

func (procPriority) ID() string   { return "proc-priority-high" }
func (procPriority) Kind() Kind   { return Live }

func priorityOf(pid uint32) (uint32, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(h)
	return windows.GetPriorityClass(h)
}

func setPriority(pid, class uint32) error {
	h, err := windows.OpenProcess(windows.PROCESS_SET_INFORMATION|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	return windows.SetPriorityClass(h, class)
}

func (p procPriority) State(c Context) State {
	if c.PID == 0 {
		return State{Applicable: false, Detail: "roblox not running"}
	}
	cur, err := priorityOf(c.PID)
	if err != nil {
		return State{Applicable: false, Detail: err.Error()}
	}
	return State{Applicable: true, Applied: cur == p.class, Detail: "class 0x" + strconv.FormatUint(uint64(cur), 16)}
}

func (p procPriority) Apply(c Context) (Undo, error) {
	if c.PID == 0 {
		return nil, errors.New("roblox not running")
	}
	prev, err := priorityOf(c.PID)
	if err != nil {
		return nil, err
	}
	if err := setPriority(c.PID, p.class); err != nil {
		return nil, err
	}
	return Undo{"class": strconv.FormatUint(uint64(prev), 10)}, nil
}

func (p procPriority) Revert(c Context, u Undo) error {
	if c.PID == 0 {
		return nil
	}
	prev, err := strconv.ParseUint(u["class"], 10, 32)
	if err != nil {
		prev = windows.NORMAL_PRIORITY_CLASS
	}
	return setPriority(c.PID, uint32(prev))
}

const ifeoKey = `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Image File Execution Options\RobloxPlayerBeta.exe\PerfOptions`

type ifeoPriority struct{}

func (ifeoPriority) ID() string { return "ifeo-priority-high" }
func (ifeoPriority) Kind() Kind { return Relaunch }

func (ifeoPriority) State(Context) State {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, ifeoKey, registry.QUERY_VALUE)
	if err != nil {
		return State{Applicable: true, Applied: false, Detail: "not set"}
	}
	defer k.Close()
	v, _, err := k.GetIntegerValue("CpuPriorityClass")
	return State{Applicable: true, Applied: err == nil && v == 3, Detail: "CpuPriorityClass=" + strconv.FormatUint(v, 10)}
}

func (ifeoPriority) Apply(Context) (Undo, error) {
	k, existed, err := registry.CreateKey(registry.LOCAL_MACHINE, ifeoKey, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()
	u := Undo{"existed": strconv.FormatBool(existed)}
	if v, _, err := k.GetIntegerValue("CpuPriorityClass"); err == nil {
		u["prev"] = strconv.FormatUint(v, 10)
	}
	return u, k.SetDWordValue("CpuPriorityClass", 3)
}

func (ifeoPriority) Revert(_ Context, u Undo) error {
	if prev, ok := u["prev"]; ok {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, ifeoKey, registry.SET_VALUE)
		if err != nil {
			return err
		}
		defer k.Close()
		v, _ := strconv.ParseUint(prev, 10, 32)
		return k.SetDWordValue("CpuPriorityClass", uint32(v))
	}
	if u["existed"] == "true" {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, ifeoKey, registry.SET_VALUE)
		if err != nil {
			return err
		}
		defer k.Close()
		return k.DeleteValue("CpuPriorityClass")
	}
	return registry.DeleteKey(registry.LOCAL_MACHINE, ifeoKey)
}

func init() {
	Register(procPriority{class: windows.HIGH_PRIORITY_CLASS})
	Register(ifeoPriority{})
}
