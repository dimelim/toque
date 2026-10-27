package session

import (
	"context"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"toque/internal/roblox"
	"toque/internal/tweaks"
)

var (
	kernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	procSetProcessInfo       = kernel32.NewProc("SetProcessInformation")
	procCreateMutexW         = kernel32.NewProc("CreateMutexW")
)

const (
	processPowerThrottling                    = 4
	powerThrottlingCurrentVersion             = 1
	powerThrottlingExecutionSpeed             = 0x1
	powerThrottlingIgnoreTimerResolution      = 0x4
)

type powerThrottlingState struct {
	Version     uint32
	ControlMask uint32
	StateMask   uint32
}

func keepTimerRequestsHonored() {
	st := powerThrottlingState{
		Version:     powerThrottlingCurrentVersion,
		ControlMask: powerThrottlingIgnoreTimerResolution | powerThrottlingExecutionSpeed,
		StateMask:   0,
	}
	procSetProcessInfo.Call(uintptr(windows.CurrentProcess()), processPowerThrottling, uintptr(unsafe.Pointer(&st)), unsafe.Sizeof(st))
}

func Exclusive(name string) (bool, func()) {
	n, _ := windows.UTF16PtrFromString(name)
	h, _, err := procCreateMutexW.Call(0, 1, uintptr(unsafe.Pointer(n)))
	if h == 0 {
		return false, func() {}
	}
	if err == windows.ERROR_ALREADY_EXISTS {
		windows.CloseHandle(windows.Handle(h))
		return false, func() {}
	}
	return true, func() { windows.CloseHandle(windows.Handle(h)) }
}

type Status struct {
	Active  bool
	Applied []string
	Since   time.Time
}

type Controller struct {
	mu      sync.Mutex
	status  Status
	applied map[string]tweaks.Undo
	live    []string
	OnChange func(Status)
}

func NewController(liveTweaks []string) *Controller {
	keepTimerRequestsHonored()
	return &Controller{live: liveTweaks, applied: map[string]tweaks.Undo{}}
}

func (c *Controller) Status() Status {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}

func (c *Controller) set(s Status) {
	c.mu.Lock()
	c.status = s
	c.mu.Unlock()
	if c.OnChange != nil {
		c.OnChange(s)
	}
}

func (c *Controller) apply(pid uint32) {
	ctx := tweaks.Context{PID: pid}
	var names []string
	for _, id := range c.live {
		t, err := tweaks.Get(id)
		if err != nil || t.Kind() != tweaks.Live {
			continue
		}
		undo, err := t.Apply(ctx)
		if err != nil {
			continue
		}
		c.applied[id] = undo
		names = append(names, id)
	}
	c.set(Status{Active: true, Applied: names, Since: time.Now()})
}

func (c *Controller) release(pid uint32) {
	ctx := tweaks.Context{PID: pid}
	for id, undo := range c.applied {
		if t, err := tweaks.Get(id); err == nil {
			t.Revert(ctx, undo)
		}
		delete(c.applied, id)
	}
	c.set(Status{})
}

func (c *Controller) Run(ctx context.Context) {
	var pid uint32
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for {
		pids := roblox.Running()
		switch {
		case len(pids) > 0 && pid == 0:
			pid = pids[0]
			c.apply(pid)
		case len(pids) == 0 && pid != 0:
			c.release(pid)
			pid = 0
		case len(pids) > 0 && pids[0] != pid:
			c.release(pid)
			pid = pids[0]
			c.apply(pid)
		}
		select {
		case <-ctx.Done():
			if pid != 0 {
				c.release(pid)
			}
			return
		case <-tick.C:
		}
	}
}
