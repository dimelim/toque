package overlay

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var keyColor = rgb(255, 0, 255)

type rgbc struct{ r, g, b uint8 }

func (c rgbc) ref() uintptr { return rgb(uint32(c.r), uint32(c.g), uint32(c.b)) }

func lerp(a, b rgbc, f float64) rgbc {
	if f < 0 {
		f = 0
	}
	if f > 1 {
		f = 1
	}
	return rgbc{
		uint8(float64(a.r) + (float64(b.r)-float64(a.r))*f),
		uint8(float64(a.g) + (float64(b.g)-float64(a.g))*f),
		uint8(float64(a.b) + (float64(b.b)-float64(a.b))*f),
	}
}

var (
	cPanel    = rgbc{13, 17, 11}
	cEdge     = rgbc{46, 58, 42}
	cKeyIdle  = rgbc{30, 38, 28}
	cKeyLit   = rgbc{182, 242, 60}
	cTxtIdle  = rgbc{156, 170, 154}
	cTxtLit   = rgbc{12, 16, 8}
	cWhite    = rgbc{238, 244, 236}
	cMuted    = rgbc{150, 166, 150}
	cMouse    = rgbc{24, 30, 22}
	fadeNanos = float64((160 * time.Millisecond).Nanoseconds())
)

type keyBox struct {
	label string
	vk    int32
	r     rect
}

type layout struct {
	w, ht int
	ping  rect
	keys  []keyBox
	mouse rect
	lmb   rect
	rmb   rect
	dot   rect
}

func panelSize(cfg Config) (int, int) {
	l := layoutOf(cfg)
	return l.w, l.ht
}

func PanelSize(cfg Config) (int, int) { return panelSize(cfg.norm()) }

func ScreenSize() (int, int) { return screenSize() }

func layoutOf(cfg Config) layout {
	s := cfg.Scale
	pad := int(14 * s)
	u := int(40 * s)
	g := int(7 * s)
	ph := 0
	if cfg.ShowPing {
		ph = int(26 * s)
	}
	top := pad + ph
	x := pad
	var l layout

	if cfg.ShowKeys {
		kbW := 3*u + 2*g
		kbTop := top
		w := keyBox{"W", 0x57, box(x+u+g, kbTop, u, u)}
		row2 := kbTop + u + g
		a := keyBox{"A", 0x41, box(x, row2, u, u)}
		sKey := keyBox{"S", 0x53, box(x+u+g, row2, u, u)}
		d := keyBox{"D", 0x44, box(x+2*(u+g), row2, u, u)}
		row3 := row2 + u + g
		shiftW := (kbW - g) / 2
		shift := keyBox{"Shift", 0xA0, box(x, row3, shiftW, u)}
		space := keyBox{"Space", 0x20, box(x+shiftW+g, row3, kbW-shiftW-g, u)}
		l.keys = []keyBox{w, a, sKey, d, shift, space}
		x += kbW + pad
	}

	if cfg.ShowMouse {
		mw := int(1.55 * float64(u))
		mh := int(2.25 * float64(u))
		mTop := top
		if cfg.ShowKeys {
			kbH := 3*u + 2*g
			mTop = top + (kbH-mh)/2
		}
		l.mouse = box(x, mTop, mw, mh)
		bh := int(0.42 * float64(mh))
		half := (mw - g) / 2
		l.lmb = box(x, mTop, half, bh)
		l.rmb = box(x+mw-half, mTop, half, bh)
		ds := int(10 * s)
		l.dot = box(x+mw/2-ds/2, mTop+bh+(mh-bh)/2-ds/2, ds, ds)
		x += mw + pad
	}

	bottom := top
	if cfg.ShowKeys {
		bottom = top + 3*u + 2*g
	}
	if cfg.ShowMouse {
		if mb := l.mouse.bottom; int(mb) > bottom {
			bottom = int(mb)
		}
	}
	l.w = x
	if l.w < pad*2+u {
		l.w = pad*2 + u
	}
	l.ht = bottom + pad
	l.ping = rect{int32(pad), int32(pad), int32(l.w - pad), int32(pad + ph)}
	return l
}

func box(x, y, w, h int) rect {
	return rect{int32(x), int32(y), int32(x + w), int32(y + h)}
}

type paintStruct struct {
	hdc       uintptr
	erase     int32
	rc        rect
	restore   int32
	incUpdate int32
	reserved  [32]byte
}

var (
	litMu sync.Mutex
	litAt = map[int32]int64{}
	lastC point
)

func fade(vk int32, down bool) float64 {
	now := time.Now().UnixNano()
	litMu.Lock()
	defer litMu.Unlock()
	if down {
		litAt[vk] = now
		return 1
	}
	last, ok := litAt[vk]
	if !ok {
		return 0
	}
	f := 1 - float64(now-last)/fadeNanos
	if f < 0 {
		return 0
	}
	return f
}

type frame struct {
	ping  float64
	keys  [6]float64
	mouse [2]float64
	dot   rect
}

func quant(f float64) float64 { return float64(int(f*16)) / 16 }

func (h *HUD) nextFrame(cfg Config, ping float64) frame {
	l := layoutOf(cfg)
	var fr frame
	if cfg.Demo && ping == 0 {
		ping = 23
	}
	fr.ping = float64(int(ping))
	for i, k := range l.keys {
		down := keyDown(k.vk) || cfg.Demo && (k.label == "W" || k.label == "S" || k.label == "Space")
		fr.keys[i] = quant(fade(k.vk, down))
	}
	if cfg.ShowMouse {
		for i, vk := range []int32{0x01, 0x02} {
			down := keyDown(vk) || cfg.Demo && vk == 0x01
			fr.mouse[i] = quant(fade(vk, down))
		}
		fr.dot = mouseDot(l.dot, cfg)
	}
	return fr
}

func (h *HUD) fonts(scale float64) (key, small, ping uintptr) {
	if h.fontScale != scale {
		h.dropFonts()
		h.fontKey = createFont(int(15*scale), true)
		h.fontSmall = createFont(int(12*scale), true)
		h.fontPing = createFont(int(16*scale), true)
		h.fontScale = scale
	}
	return h.fontKey, h.fontSmall, h.fontPing
}

func (h *HUD) dropFonts() {
	for _, f := range []uintptr{h.fontKey, h.fontSmall, h.fontPing} {
		if f != 0 {
			deleteObject.Call(f)
		}
	}
	h.fontKey, h.fontSmall, h.fontPing, h.fontScale = 0, 0, 0, 0
}

func (h *HUD) paint(hwnd windows.HWND) {
	cfg, _ := h.snapshot()
	fr := h.frame
	l := layoutOf(cfg)
	var ps paintStruct
	hdc, _, _ := beginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
	mem, _, _ := createCompatibleDC.Call(hdc)
	bmp, _, _ := createCompatibleBmp.Call(hdc, uintptr(l.w), uintptr(l.ht))
	old, _, _ := selectObject.Call(mem, bmp)
	setBkMode.Call(mem, transparentBk)

	full := rect{0, 0, int32(l.w), int32(l.ht)}
	kb, _, _ := createSolidBrush.Call(keyColor)
	fillRect.Call(mem, uintptr(unsafe.Pointer(&full)), kb)
	deleteObject.Call(kb)

	radius := int(18 * cfg.Scale)
	fillRound(mem, full, radius, cPanel, cEdge)

	fontKey, fontSmall, fontPing := h.fonts(cfg.Scale)

	if cfg.ShowPing {
		txt := "sin medir"
		col := cMuted
		if ping := fr.ping; ping > 0 {
			txt = fmt.Sprintf("%.0f ms de ping", ping)
			if ping <= 60 {
				col = cKeyLit
			} else {
				col = cWhite
			}
		}
		drawText(mem, l.ping, txt, col, fontPing, dtLeft|dtVCenter|dtSingleLine)
	}

	for i, k := range l.keys {
		f := fr.keys[i]
		fill := lerp(cKeyIdle, cKeyLit, f)
		edge := lerp(cEdge, cKeyLit, f)
		txt := lerp(cTxtIdle, cTxtLit, f)
		fillRound(mem, k.r, int(9*cfg.Scale), fill, edge)
		font := fontKey
		if len(k.label) > 1 {
			font = fontSmall
		}
		drawText(mem, k.r, k.label, txt, font, dtCenter|dtVCenter|dtSingleLine)
	}

	if cfg.ShowMouse {
		fillRound(mem, l.mouse, int(float64(l.mouse.right-l.mouse.left)*0.45), cMouse, cEdge)
		for i, r := range []rect{l.lmb, l.rmb} {
			f := fr.mouse[i]
			fillRound(mem, r, int(7*cfg.Scale), lerp(cKeyIdle, cKeyLit, f), lerp(cEdge, cKeyLit, f))
		}
		dot := fr.dot
		db, _, _ := createSolidBrush.Call(cMuted.ref())
		penH, _ := stock(nullPen)
		op, _, _ := selectObject.Call(mem, uintptr(penH))
		ob, _, _ := selectObject.Call(mem, db)
		ellipse.Call(mem, uintptr(dot.left), uintptr(dot.top), uintptr(dot.right), uintptr(dot.bottom))
		selectObject.Call(mem, op)
		selectObject.Call(mem, ob)
		deleteObject.Call(db)
	}

	bitBlt.Call(hdc, 0, 0, uintptr(l.w), uintptr(l.ht), mem, 0, 0, 0x00CC0020)

	selectObject.Call(mem, old)
	deleteObject.Call(bmp)
	deleteDC.Call(mem)
	endPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
}

func mouseDot(base rect, cfg Config) rect {
	var p point
	getCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	litMu.Lock()
	dx := float64(p.x-lastC.x) / 14
	dy := float64(p.y-lastC.y) / 14
	lastC = p
	litMu.Unlock()
	rng := 6 * cfg.Scale
	dx = clampF(dx, -rng, rng)
	dy = clampF(dy, -rng, rng)
	return rect{base.left + int32(dx), base.top + int32(dy), base.right + int32(dx), base.bottom + int32(dy)}
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func fillRound(dc uintptr, r rect, radius int, fill, border rgbc) {
	brush, _, _ := createSolidBrush.Call(fill.ref())
	pen, _, _ := createPen.Call(0, 1, border.ref())
	ob, _, _ := selectObject.Call(dc, brush)
	op, _, _ := selectObject.Call(dc, pen)
	roundRect.Call(dc, uintptr(r.left), uintptr(r.top), uintptr(r.right), uintptr(r.bottom), uintptr(radius), uintptr(radius))
	selectObject.Call(dc, ob)
	selectObject.Call(dc, op)
	deleteObject.Call(brush)
	deleteObject.Call(pen)
}

func drawText(dc uintptr, r rect, text string, color rgbc, font uintptr, flags uintptr) {
	ptr, _ := windows.UTF16PtrFromString(text)
	of, _, _ := selectObject.Call(dc, font)
	setTextColor.Call(dc, color.ref())
	rc := r
	drawTextW.Call(dc, uintptr(unsafe.Pointer(ptr)), ^uintptr(0), uintptr(unsafe.Pointer(&rc)), flags)
	selectObject.Call(dc, of)
}

func createFont(size int, bold bool) uintptr {
	weight := uintptr(400)
	if bold {
		weight = 700
	}
	face, _ := windows.UTF16PtrFromString("Segoe UI")
	r, _, _ := createFontW.Call(uintptr(-size), 0, 0, 0, weight, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(face)))
	return r
}
