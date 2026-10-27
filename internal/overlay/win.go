package overlay

import (
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32 = windows.NewLazySystemDLL("user32.dll")
	gdi32  = windows.NewLazySystemDLL("gdi32.dll")

	registerClassExW  = user32.NewProc("RegisterClassExW")
	createWindowExW   = user32.NewProc("CreateWindowExW")
	defWindowProcW    = user32.NewProc("DefWindowProcW")
	destroyWindow     = user32.NewProc("DestroyWindow")
	showWindow        = user32.NewProc("ShowWindow")
	setWindowPos      = user32.NewProc("SetWindowPos")
	setLayeredAttrs   = user32.NewProc("SetLayeredWindowAttributes")
	getMessageW       = user32.NewProc("GetMessageW")
	translateMessage  = user32.NewProc("TranslateMessage")
	dispatchMessageW  = user32.NewProc("DispatchMessageW")
	postQuitMessage   = user32.NewProc("PostQuitMessage")
	postThreadMessage = user32.NewProc("PostThreadMessageW")
	setTimer          = user32.NewProc("SetTimer")
	killTimer         = user32.NewProc("KillTimer")
	beginPaint        = user32.NewProc("BeginPaint")
	endPaint          = user32.NewProc("EndPaint")
	invalidateRect    = user32.NewProc("InvalidateRect")
	getClientRect     = user32.NewProc("GetClientRect")
	fillRect          = user32.NewProc("FillRect")
	getAsyncKeyState  = user32.NewProc("GetAsyncKeyState")
	getCursorPos      = user32.NewProc("GetCursorPos")
	getSystemMetrics  = user32.NewProc("GetSystemMetrics")
	drawTextW         = user32.NewProc("DrawTextW")
	getDC             = user32.NewProc("GetDC")
	releaseDC         = user32.NewProc("ReleaseDC")

	createCompatibleDC  = gdi32.NewProc("CreateCompatibleDC")
	createCompatibleBmp = gdi32.NewProc("CreateCompatibleBitmap")
	selectObject        = gdi32.NewProc("SelectObject")
	deleteObject        = gdi32.NewProc("DeleteObject")
	deleteDC            = gdi32.NewProc("DeleteDC")
	bitBlt              = gdi32.NewProc("BitBlt")
	createSolidBrush    = gdi32.NewProc("CreateSolidBrush")
	createPen           = gdi32.NewProc("CreatePen")
	roundRect           = gdi32.NewProc("RoundRect")
	ellipse             = gdi32.NewProc("Ellipse")
	setBkMode           = gdi32.NewProc("SetBkMode")
	setTextColor        = gdi32.NewProc("SetTextColor")
	createFontW         = gdi32.NewProc("CreateFontW")
	getStockObject      = gdi32.NewProc("GetStockObject")
)

const (
	wmDestroy = 0x0002
	wmTimer   = 0x0113
	wmPaint   = 0x000F
	wmQuit    = 0x0012
	wmNCHit   = 0x0084

	wsPopup       = 0x80000000
	wsExLayered   = 0x00080000
	wsExTransp    = 0x00000020
	wsExTopmost   = 0x00000008
	wsExToolWin   = 0x00000080
	wsExNoActive  = 0x08000000
	lwaColorKey   = 0x00000001
	lwaAlpha      = 0x00000002
	swpNoSize     = 0x0001
	swpNoActivate = 0x0010
	swpNoMove     = 0x0002
	hwndTopmost   = ^uintptr(0) // (HWND)-1
	htCaption     = 2
	swHide        = 0
	swShowNA      = 8
	frameMs       = 16
	hiddenMs      = 250
	nullPen       = 8
	transparentBk = 1
	dtCenter      = 0x00000001
	dtVCenter     = 0x00000004
	dtSingleLine  = 0x00000020
	dtLeft        = 0x00000000
)

type rect struct{ left, top, right, bottom int32 }

type point struct{ x, y int32 }

type msg struct {
	hwnd    windows.HWND
	message uint32
	wparam  uintptr
	lparam  uintptr
	time    uint32
	pt      point
}

type wndClassEx struct {
	size       uint32
	style      uint32
	wndProc    uintptr
	clsExtra   int32
	wndExtra   int32
	instance   windows.Handle
	icon       windows.Handle
	cursor     windows.Handle
	background windows.Handle
	menuName   *uint16
	className  *uint16
	iconSm     windows.Handle
}

func rgb(r, g, b uint32) uintptr { return uintptr(r | g<<8 | b<<16) }

func keyDown(vk int32) bool {
	r, _, _ := getAsyncKeyState.Call(uintptr(vk))
	return r&0x8000 != 0
}

func (h *HUD) loop(ready chan struct{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	h.mu.Lock()
	h.tid = windows.GetCurrentThreadId()
	h.mu.Unlock()

	className, _ := windows.UTF16PtrFromString("ToqueHudClass")
	proc := syscall.NewCallback(h.wndProc)
	wc := wndClassEx{
		style:     0x0002 | 0x0001, // CS_HREDRAW | CS_VREDRAW
		wndProc:   proc,
		className: className,
	}
	wc.size = uint32(unsafe.Sizeof(wc))
	registerClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	cfg, _ := h.snapshot()
	w, ht := panelSize(cfg)
	ex := uintptr(wsExLayered | wsExTopmost | wsExToolWin | wsExNoActive)
	if !cfg.Move {
		ex |= wsExTransp
	}
	title, _ := windows.UTF16PtrFromString("Toque HUD")
	hwnd, _, _ := createWindowExW.Call(
		ex,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		uintptr(wsPopup),
		uintptr(cfg.X), uintptr(cfg.Y), uintptr(w), uintptr(ht),
		0, 0, 0, 0,
	)
	if hwnd == 0 {
		h.mu.Lock()
		h.run = false
		h.mu.Unlock()
		close(ready)
		return
	}
	h.mu.Lock()
	h.hwnd = windows.HWND(hwnd)
	h.mu.Unlock()

	alpha := uintptr(cfg.Opacity) * 255 / 100
	setLayeredAttrs.Call(hwnd, keyColor, alpha, lwaColorKey|lwaAlpha)
	setWindowPos.Call(hwnd, hwndTopmost, 0, 0, 0, 0, swpNoMove|swpNoSize|swpNoActivate)
	h.shown, h.drawn = false, false
	setTimer.Call(hwnd, 1, hiddenMs, 0)
	close(ready)

	var m msg
	for {
		r, _, _ := getMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&m)))
		dispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	h.mu.Lock()
	if h.hwnd != 0 {
		destroyWindow.Call(uintptr(h.hwnd))
		h.hwnd = 0
	}
	h.run = false
	h.tid = 0
	h.mu.Unlock()
}

func (h *HUD) wndProc(hwnd windows.HWND, message uint32, wparam, lparam uintptr) uintptr {
	switch message {
	case wmTimer:
		h.onTick(hwnd)
		return 0
	case wmPaint:
		h.paint(hwnd)
		return 0
	case wmNCHit:
		cfg, _ := h.snapshot()
		if cfg.Move {
			return htCaption
		}
	case wmDestroy:
		killTimer.Call(uintptr(hwnd), 1)
		h.dropFonts()
		postQuitMessage.Call(0)
		return 0
	}
	r, _, _ := defWindowProcW.Call(uintptr(hwnd), uintptr(message), wparam, lparam)
	return r
}

func (h *HUD) onTick(hwnd windows.HWND) {
	cfg, ping := h.snapshot()
	visible := cfg.Move || cfg.Demo || robloxInFront()
	if visible != h.shown {
		h.shown, h.drawn = visible, false
		if visible {
			setWindowPos.Call(uintptr(hwnd), hwndTopmost, 0, 0, 0, 0, swpNoMove|swpNoSize|swpNoActivate)
			showWindow.Call(uintptr(hwnd), swShowNA)
			setTimer.Call(uintptr(hwnd), 1, frameMs, 0)
		} else {
			showWindow.Call(uintptr(hwnd), swHide)
			setTimer.Call(uintptr(hwnd), 1, hiddenMs, 0)
		}
	}
	if !visible {
		return
	}
	fr := h.nextFrame(cfg, ping)
	if !h.drawn || fr != h.frame {
		h.frame, h.drawn = fr, true
		invalidateRect.Call(uintptr(hwnd), 0, 0)
	}
	h.tick++
	if h.tick%120 == 0 {
		setWindowPos.Call(uintptr(hwnd), hwndTopmost, 0, 0, 0, 0, swpNoMove|swpNoSize|swpNoActivate)
	}
}

func robloxInFront() bool {
	fg := windows.GetForegroundWindow()
	if fg == 0 {
		return false
	}
	buf := make([]uint16, 32)
	n, _ := windows.GetClassName(fg, &buf[0], int32(len(buf)))
	return windows.UTF16ToString(buf[:n]) == "WINDOWSCLIENT"
}

func stock(id int) (windows.Handle, uintptr) {
	r, _, _ := getStockObject.Call(uintptr(id))
	return windows.Handle(r), r
}

func screenSize() (int, int) {
	w, _, _ := getSystemMetrics.Call(0)
	ht, _, _ := getSystemMetrics.Call(1)
	return int(w), int(ht)
}
