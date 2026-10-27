package app

import (
	"context"
	"errors"
	"sync"
	"time"

	"toque/internal/bench"
	"toque/internal/catalog"
	"toque/internal/roblox"
	"toque/internal/tweaks"
	"toque/internal/winx"
)

const Version = "1.0.0"

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
	live       liveSession
}

func New(presentMon string, emit Emitter) *Service {
	return &Service{presentMon: presentMon, emit: emit}
}

var ErrBusy = errors.New("toque ya está ejecutando otra tarea")

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

type LiveState struct {
	Agent  bool     `json:"agent"`
	Active bool     `json:"active"`
	Tweaks []string `json:"tweaks"`
}

type Overview struct {
	Version   string      `json:"version"`
	Elevated  bool        `json:"elevated"`
	Roblox    RobloxState `json:"roblox"`
	System    SystemInfo  `json:"system"`
	Applied   int         `json:"applied"`
	Proven    int         `json:"proven"`
	LastRun   *RunRecord  `json:"lastRun"`
	Busy      string      `json:"busy"`
	UptimeMin int         `json:"uptimeMin"`
	Live      LiveState   `json:"live"`
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
	live := catalog.LiveProven()
	o.Live = LiveState{Agent: AgentRunning(), Active: o.Roblox.Running && len(live) > 0, Tweaks: catalog.Titles(live)}
	o.Applied = len(tweaks.OpenJournal().Applied)
	for _, e := range catalog.Catalog {
		if e.Verdict == catalog.VerdictProven {
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

func currentSession(inst roblox.Install) (roblox.Session, error) {
	pids := roblox.Running()
	if len(pids) == 0 {
		return roblox.Session{}, errors.New("roblox is not running")
	}
	log := roblox.NewestPlayerLog(inst.LogDir(), time.Now().Add(-12*time.Hour))
	if log == "" {
		return roblox.Session{PID: pids[0]}, errors.New("no log")
	}
	sess := roblox.Session{PID: pids[0], LogPath: log}
	for _, ev := range roblox.ScanFile(log) {
		switch ev.Kind {
		case roblox.EventJoining:
			sess.JobID, sess.PlaceID, sess.ServerIP = ev.JobID, ev.PlaceID, ""
		case roblox.EventUdmux:
			sess.ServerIP, sess.ServerPort = ev.IP, ev.Port
		case roblox.EventDisconnected:
			sess.ServerIP = ""
		}
	}
	if sess.ServerIP == "" {
		return sess, errors.New("not in game")
	}
	return sess, nil
}

func (s *Service) Play() error {
	inst, err := roblox.Find()
	if err != nil {
		return err
	}
	_, err = winx.LaunchAsUser(inst.Player, roblox.JoinURL(roblox.RealisticStreetSoccer, ""), inst.VersionDir)
	return err
}
