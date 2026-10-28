package overlay

import (
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

type Config struct {
	Enabled   bool    `json:"enabled"`
	X         int     `json:"x"`
	Y         int     `json:"y"`
	Scale     float64 `json:"scale"`
	Opacity   int     `json:"opacity"`
	ShowKeys  bool    `json:"showKeys"`
	ShowMouse bool    `json:"showMouse"`
	ShowPing  bool    `json:"showPing"`
	Move      bool    `json:"move"`
	Demo      bool    `json:"demo"`
}

func Defaults() Config {
	return Config{X: 40, Y: 240, Scale: 1, Opacity: 90, ShowKeys: true, ShowMouse: true, ShowPing: true}
}

func (c Config) norm() Config {
	if c.Scale < 0.6 {
		c.Scale = 0.6
	}
	if c.Scale > 2.2 {
		c.Scale = 2.2
	}
	if c.Opacity < 30 {
		c.Opacity = 30
	}
	if c.Opacity > 100 {
		c.Opacity = 100
	}
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
	return c
}

type HUD struct {
	mu      sync.Mutex
	cfg     Config
	ping    float64
	preview bool
	run     bool
	tid     uint32
	tick    int
	hwnd    windows.HWND

	frame     frame
	drawn     bool
	shown     bool
	fontScale float64
	fontKey   uintptr
	fontSmall uintptr
	fontPing  uintptr
}

func New() *HUD { return &HUD{cfg: Defaults()} }

func (h *HUD) SetPing(ms float64) {
	h.mu.Lock()
	h.ping = ms
	h.mu.Unlock()
}

func (h *HUD) SetPreview(on bool) {
	h.mu.Lock()
	h.preview = on
	h.mu.Unlock()
}

func (h *HUD) previewing() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.preview
}

func (h *HUD) MoveTo(x, y int) Config {
	h.mu.Lock()
	h.cfg.X, h.cfg.Y = x, y
	h.cfg = h.cfg.norm()
	cfg, hwnd := h.cfg, h.hwnd
	h.mu.Unlock()
	if hwnd != 0 {
		setWindowPos.Call(uintptr(hwnd), hwndTopmost, uintptr(cfg.X), uintptr(cfg.Y), 0, 0, swpNoSize|swpNoActivate)
	}
	return cfg
}

func (h *HUD) Config() Config {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.cfg
}

func (h *HUD) Apply(cfg Config) {
	cfg = cfg.norm()
	h.mu.Lock()
	h.cfg = cfg
	running := h.run
	h.mu.Unlock()
	if running {
		h.stop()
	}
	if cfg.Enabled {
		h.start()
	}
}

func (h *HUD) snapshot() (Config, float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.cfg, h.ping
}

func (h *HUD) start() {
	h.mu.Lock()
	if h.run {
		h.mu.Unlock()
		return
	}
	h.run = true
	h.mu.Unlock()
	ready := make(chan struct{})
	go h.loop(ready)
	<-ready
}

func (h *HUD) stop() {
	h.mu.Lock()
	tid, running := h.tid, h.run
	h.mu.Unlock()
	if !running || tid == 0 {
		return
	}
	postThreadMessage.Call(uintptr(tid), wmQuit, 0, 0)
	for i := 0; i < 50; i++ {
		h.mu.Lock()
		done := !h.run
		h.mu.Unlock()
		if done {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
