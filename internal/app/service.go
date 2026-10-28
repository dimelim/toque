package app

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"toque/internal/bench"
	"toque/internal/catalog"
	"toque/internal/overlay"
	"toque/internal/roblox"
	"toque/internal/tweaks"
	"toque/internal/winx"
)

const Version = "1.16.0"

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
	hud        *overlay.HUD
	sess       *session
	calm       atomic.Bool
}

func New(presentMon string, emit Emitter) *Service {
	s := &Service{presentMon: presentMon, emit: emit, upd: UpdateInfo{Current: Version, State: "idle"}, hud: overlay.New(), sess: loadSession()}
	s.hud.Apply(loadHudConfig())
	return s
}

var ErrBusy = errors.New("toque ya estÃ¡ ejecutando otra tarea")

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
	measuring := task == "verify" || task == "benchmark" || task == "diagnose"
	if measuring {
		s.pauseSession()
	}
	return ctx, func() {
		cancel()
		s.mu.Lock()
		s.busy, s.cancel = "", nil
		s.mu.Unlock()
		if measuring {
			s.resumeSession()
		}
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

func (s *Service) launchInstall() (roblox.Install, error) {
	inst, err := roblox.Find()
	if err != nil {
		return roblox.Install{}, err
	}
	tweaks.Resync(tweaks.Context{Install: inst})
	s.prelaunch()
	return inst, nil
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

func robloxState() RobloxState {
	var r RobloxState
	inst, err := roblox.Find()
	if err != nil {
		return r
	}
	r.Installed, r.Version = true, inst.Version
	if sess, err := currentSession(inst); err == nil {
		r.Running = true
		r.InGame = sess.ServerIP != ""
		r.PlaceID = sess.PlaceID
		r.JobID = sess.JobID
		r.ServerIP = sess.ServerIP
	} else {
		r.Running = len(roblox.Running()) > 0
	}
	return r
}

func (s *Service) Overview() Overview {
	systemOnce.Do(func() { systemInfo = ReadSystem() })
	o := Overview{Version: Version, Elevated: winx.IsElevated(), System: systemInfo, UptimeMin: int(uptime().Minutes()), Roblox: robloxState()}
	entries, _ := tweaks.Entries()
	o.Applied = len(entries)
	o.Tested = len(catalog.Catalog)
	for _, e := range catalog.Catalog {
		if e.Verdict == catalog.VerdictProven || e.Verdict == catalog.VerdictTradeoff {
			o.Proven++
		}
	}
	o.LastRun = latestRun()
	s.mu.Lock()
	o.Busy = s.busy
	s.mu.Unlock()
	s.feedPing(o)
	return o
}

func (s *Service) feedPing(o Overview) {
	ping := 0.0
	if o.Roblox.InGame && o.Roblox.ServerIP != "" {
		if rtt, ok := knownDatacenter(o.Roblox.ServerIP); ok {
			ping = rtt
		}
	}
	if ping == 0 {
		if best, ok := bestDatacenter(); ok {
			ping = best
		}
	}
	s.hud.SetPing(ping)
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
	throttler().Recover()
	restoreSystem()
	tweaks.Revert(tweaks.Context{}, backgroundTweak{})
	for _, t := range systemTweaks {
		tweaks.Revert(tweaks.Context{}, t)
	}
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
