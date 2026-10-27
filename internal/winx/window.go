package winx

import (
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procShowWindow          = user32.NewProc("ShowWindow")
	procIsIconic            = user32.NewProc("IsIconic")
	procKeybdEvent          = user32.NewProc("keybd_event")
	procGetLastInputInfo    = user32.NewProc("GetLastInputInfo")
	procGetTickCount        = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetTickCount")
)

func MainWindow(pid uint32) windows.HWND {
	var found windows.HWND
	cb := syscall.NewCallback(func(h windows.HWND, _ uintptr) uintptr {
		var wpid uint32
		windows.GetWindowThreadProcessId(h, &wpid)
		if wpid != pid || !windows.IsWindowVisible(h) {
			return 1
		}
		buf := make([]uint16, 64)
		n, _ := windows.GetClassName(h, &buf[0], int32(len(buf)))
		if windows.UTF16ToString(buf[:n]) == "WINDOWSCLIENT" {
			found = h
			return 0
		}
		return 1
	})
	windows.EnumWindows(cb, nil)
	return found
}

func Focus(h windows.HWND) bool {
	if h == 0 {
		return false
	}
	if r, _, _ := procIsIconic.Call(uintptr(h)); r != 0 {
		procShowWindow.Call(uintptr(h), 9)
	}
	procKeybdEvent.Call(0x12, 0, 0, 0)
	procKeybdEvent.Call(0x12, 0, 2, 0)
	procSetForegroundWindow.Call(uintptr(h))
	time.Sleep(300 * time.Millisecond)
	return windows.GetForegroundWindow() == h
}

var (
	procFindWindowW  = user32.NewProc("FindWindowW")
	procPostMessageW = user32.NewProc("PostMessageW")
)

func CloseWindowByClass(class string) bool {
	c, _ := windows.UTF16PtrFromString(class)
	h, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(c)), 0)
	if h == 0 {
		return false
	}
	r, _, _ := procPostMessageW.Call(h, 0x0010, 0, 0)
	return r != 0
}

func IsForeground(h windows.HWND) bool {
	return h != 0 && windows.GetForegroundWindow() == h
}

func IdleTime() time.Duration {
	var info struct {
		Size uint32
		Time uint32
	}
	info.Size = uint32(unsafe.Sizeof(info))
	if r, _, _ := procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&info))); r == 0 {
		return 0
	}
	now, _, _ := procGetTickCount.Call()
	return time.Duration(uint32(now)-info.Time) * time.Millisecond
}
