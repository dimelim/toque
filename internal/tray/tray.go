package tray

import (
	"errors"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	shell32                 = windows.NewLazySystemDLL("shell32.dll")
	procRegisterClassExW    = user32.NewProc("RegisterClassExW")
	procCreateWindowExW     = user32.NewProc("CreateWindowExW")
	procDefWindowProcW      = user32.NewProc("DefWindowProcW")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessageW    = user32.NewProc("DispatchMessageW")
	procPostQuitMessage     = user32.NewProc("PostQuitMessage")
	procPostMessageW        = user32.NewProc("PostMessageW")
	procRegisterWindowMsgW  = user32.NewProc("RegisterWindowMessageW")
	procCreatePopupMenu     = user32.NewProc("CreatePopupMenu")
	procAppendMenuW         = user32.NewProc("AppendMenuW")
	procTrackPopupMenu      = user32.NewProc("TrackPopupMenu")
	procDestroyMenu         = user32.NewProc("DestroyMenu")
	procGetCursorPos        = user32.NewProc("GetCursorPos")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procDestroyWindow       = user32.NewProc("DestroyWindow")
	procShellNotifyIconW    = shell32.NewProc("Shell_NotifyIconW")
	procExtractIconExW      = shell32.NewProc("ExtractIconExW")
)

const (
	wmDestroy       = 0x0002
	wmClose         = 0x0010
	wmNull          = 0x0000
	wmApp           = 0x8000
	wmTray          = wmApp + 1
	wmRefresh       = wmApp + 2
	wmLButtonUp     = 0x0202
	wmLButtonDblClk = 0x0203
	wmRButtonUp     = 0x0205
	wmContextMenu   = 0x007B
	nimAdd          = 0
	nimModify       = 1
	nimDelete       = 2
	nifMessage      = 0x1
	nifIcon         = 0x2
	nifTip          = 0x4
	mfString        = 0x0
	mfGrayed        = 0x1
	mfSeparator     = 0x800
	tpmReturnCmd    = 0x0100
	tpmRightButton  = 0x0002
	tpmNoNotify     = 0x0080
	cmdOpen         = 1
	cmdQuit         = 2
)

type notifyIconData struct {
	CbSize           uint32
	HWnd             windows.HWND
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            windows.Handle
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         windows.GUID
	HBalloonIcon     windows.Handle
}

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   windows.Handle
	Icon       windows.Handle
	Cursor     windows.Handle
	Background windows.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     windows.Handle
}

type msg struct {
	HWnd    windows.HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

type Tray struct {
	OnOpen   func()
	OnQuit   func()
	Status   func() string
	Tooltip  func() string
	hwnd     windows.HWND
	icon     windows.Handle
	recreate uint32
}

var current *Tray

func loadIcon(path string) windows.Handle {
	p, _ := windows.UTF16PtrFromString(path)
	var small windows.Handle
	procExtractIconExW.Call(uintptr(unsafe.Pointer(p)), 0, 0, uintptr(unsafe.Pointer(&small)), 1)
	return small
}

func (t *Tray) data() notifyIconData {
	d := notifyIconData{HWnd: t.hwnd, UID: 1, UFlags: nifMessage | nifIcon | nifTip, UCallbackMessage: wmTray, HIcon: t.icon}
	d.CbSize = uint32(unsafe.Sizeof(d))
	tip := "Toque"
	if t.Tooltip != nil {
		tip = t.Tooltip()
	}
	u, _ := windows.UTF16FromString(tip)
	copy(d.SzTip[:len(d.SzTip)-1], u)
	return d
}

func (t *Tray) notify(action uintptr) {
	d := t.data()
	procShellNotifyIconW.Call(action, uintptr(unsafe.Pointer(&d)))
}

func (t *Tray) Refresh() {
	if t.hwnd != 0 {
		procPostMessageW.Call(uintptr(t.hwnd), wmRefresh, 0, 0)
	}
}

func (t *Tray) Quit() {
	if t.hwnd != 0 {
		procPostMessageW.Call(uintptr(t.hwnd), wmClose, 0, 0)
	}
}

func (t *Tray) menu() {
	m, _, _ := procCreatePopupMenu.Call()
	if m == 0 {
		return
	}
	defer procDestroyMenu.Call(m)
	add := func(flags uintptr, id uintptr, text string) {
		p, _ := windows.UTF16PtrFromString(text)
		procAppendMenuW.Call(m, flags, id, uintptr(unsafe.Pointer(p)))
	}
	if t.Status != nil {
		add(mfString|mfGrayed, 0, t.Status())
		procAppendMenuW.Call(m, mfSeparator, 0, 0)
	}
	add(mfString, cmdOpen, "Abrir Toque")
	add(mfString, cmdQuit, "Salir")
	var pt struct{ X, Y int32 }
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procSetForegroundWindow.Call(uintptr(t.hwnd))
	cmd, _, _ := procTrackPopupMenu.Call(m, tpmReturnCmd|tpmRightButton|tpmNoNotify, uintptr(pt.X), uintptr(pt.Y), 0, uintptr(t.hwnd), 0)
	procPostMessageW.Call(uintptr(t.hwnd), wmNull, 0, 0)
	switch cmd {
	case cmdOpen:
		if t.OnOpen != nil {
			go t.OnOpen()
		}
	case cmdQuit:
		t.Quit()
	}
}

func wndProc(hwnd windows.HWND, message uint32, wparam, lparam uintptr) uintptr {
	t := current
	switch {
	case t == nil:
	case message == wmTray:
		switch lparam & 0xFFFF {
		case wmLButtonDblClk:
			if t.OnOpen != nil {
				go t.OnOpen()
			}
		case wmRButtonUp, wmContextMenu, wmLButtonUp:
			t.menu()
		}
		return 0
	case message == wmRefresh:
		t.notify(nimModify)
		return 0
	case t.recreate != 0 && message == t.recreate:
		t.notify(nimAdd)
		return 0
	case message == wmClose:
		t.notify(nimDelete)
		procDestroyWindow.Call(uintptr(hwnd))
		return 0
	case message == wmDestroy:
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(message), wparam, lparam)
	return r
}

func (t *Tray) Run(iconPath string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	current = t
	var inst windows.Handle
	windows.GetModuleHandleEx(0, nil, &inst)
	cls, _ := windows.UTF16PtrFromString("ToqueAgentTray")
	wc := wndClassEx{WndProc: syscall.NewCallback(wndProc), Instance: inst, ClassName: cls}
	wc.Size = uint32(unsafe.Sizeof(wc))
	if r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		return errors.New("RegisterClassEx failed")
	}
	title, _ := windows.UTF16PtrFromString("Toque")
	h, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(title)), 0, 0, 0, 0, 0, 0, 0, uintptr(inst), 0)
	if h == 0 {
		return errors.New("CreateWindowEx failed")
	}
	t.hwnd = windows.HWND(h)
	tb, _ := windows.UTF16PtrFromString("TaskbarCreated")
	r, _, _ := procRegisterWindowMsgW.Call(uintptr(unsafe.Pointer(tb)))
	t.recreate = uint32(r)
	t.icon = loadIcon(iconPath)
	t.notify(nimAdd)
	var m msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	if t.OnQuit != nil {
		t.OnQuit()
	}
	return nil
}
