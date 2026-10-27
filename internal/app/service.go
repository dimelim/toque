package app

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"toque/internal/bench"
	"toque/internal/catalog"
	"toque/internal/roblox"
	"toque/internal/tweaks"
	"toque/internal/winx"
)

const Version = "1.11.1"

type Emitter func(name string, data any)

type Progress struct {
	Task    string  `json:"task"`
	Step    string  `json:"step"`
	Pct     float64 `json:"pct"`
	Message string  `json:"message"`
}

type Service struct {
	emit       Emitter
	presentMon string
	mu         sync.Mutex
	busy       string
	cancel     context.CancelFunc
	updMu      sync.Mutex
	upd        UpdateInfo
	checking   atomic.Bool
	watchMu    sync.Mutex
	watch      WatchInfo
	watchStop  context.CancelFunc
	watchGen   int
}

func New(presentMon string, emit Emitter) *Service {
	return &Service{presentMon: presentMon, emit: emit, upd: UpdateInfo{Current: Version, State: "idle"}}
}

var ErrBusy = errors.New("toque ya está ejecutando otra tarea")

func (s *Service) idle() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.busy != "" {
		return ErrBusy
	}
	return nil
}

func (s *Service) begin(task string) (context.Context, func(), error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.busy != "" {
		return nil, nil, ErrBusy
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.busy, s.cancel = task, cancel
	return ctx, func() {
		cancel()
		s.mu.Lock()
		s.busy, s.cancel = "", nil
		s.mu.Unlock()
		s.progress(task, "done", 1, "")
	}, nil
}

func (s *Service) progress(task, step string, pct float64, msg string) {
	if s.emit != nil {
		s.emit("progress", Progress{Task: task, Step: step, Pct: pct, Message: msg})
	}
}

func (s *Service) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Service) env() (bench.Env, error) {
	inst, err := roblox.Find()
	if err != nil {
		return bench.Env{}, errors.New("no encuentro Roblox instalado en este usuario")
	}
	return bench.Env{PresentMon: s.presentMon, Install: inst, Dir: winx.DataDir("runs")}, nil
}

type RobloxState struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
	Running   bool   `json:"running"`
	InGame    bool   `json:"inGame"`
	PlaceID   int64  `json:"placeId"`
	JobID     string `json:"jobId"`
	ServerIP  string `json:"serverIp"`
}

type Overview struct {
	Version   string      `json:"version"`
	Elevated  bool        `json:"elevated"`
	Roblox    RobloxState `json:"roblox"`
	System    SystemInfo  `json:"system"`
	Applied   int         `json:"applied"`
	Proven    int         `json:"proven"`
	Tested    int         `json:"tested"`
	LastRun   *RunRecord  `json:"lastRun"`
	Busy      string      `json:"busy"`
	UptimeMin int         `json:"uptimeMin"`
}

var (
	systemOnce sync.Once
	systemInfo SystemInfo
)

func (s *Service) Overview() Overview {
	systemOnce.Do(func() { systemInfo = ReadSystem() })
	o := Overview{Version: Version, Elevated: winx.IsElevated(), System: systemInfo, UptimeMin: int(uptime().Minutes())}
	if inst, err := roblox.Find(); err == nil {
		o.Roblox.Installed = true
		o.Roblox.Version = inst.Version
		if sess, err := currentSession(inst); err == nil {
			o.Roblox.Running = true
			o.Roblox.InGame = sess.ServerIP != ""
			o.Roblox.PlaceID = sess.PlaceID
			o.Roblox.JobID = sess.JobID
			o.Roblox.ServerIP = sess.ServerIP
		} else {
			o.Roblox.Running = len(roblox.Running()) > 0
		}
	}
	entries, _ := tweaks.Entries()
	o.Applied = len(entries)
	o.Tested = len(catalog.Catalog)
	for _, e := range catalog.Catalog {
		if e.Verdict == catalog.VerdictProven || e.Verdict == catalog.VerdictTradeoff {
			o.Proven++
		}
	}
	if runs := LoadRuns(); len(runs) > 0 {
		o.LastRun = &runs[0]
	}
	s.mu.Lock()
	o.Busy = s.busy
	s.mu.Unlock()
	return o
}

var logTail roblox.Tail

func currentSession(inst roblox.Install) (roblox.Session, error) {
	pids := roblox.Running()
	if len(pids) == 0 {
		return roblox.Session{}, errors.New("roblox is not running")
	}
	log := roblox.NewestPlayerLog(inst.LogDir(), time.Now().Add(-12*time.Hour))
	if log == "" {
		return roblox.Session{PID: pids[0]}, errors.New("no log")
	}
	st := logTail.Read(log)
	sess := roblox.Session{PID: pids[0], LogPath: log, JobID: st.JobID, PlaceID: st.PlaceID, ServerIP: st.ServerIP, ServerPort: st.ServerPort}
	if sess.ServerIP == "" {
		return sess, errors.New("not in game")
	}
	return sess, nil
}

func (s *Service) Recover() {
	inst, err := roblox.Find()
	if err != nil {
		return
	}
	bench.RecoverInterrupted(inst)
	if len(roblox.Running()) > 0 {
		return
	}
	tweaks.Resync(tweaks.Context{Install: inst})
}
