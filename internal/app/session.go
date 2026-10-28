package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"toque/internal/eco"
	"toque/internal/i18n"
	"toque/internal/roblox"
	"toque/internal/sysmon"
	"toque/internal/winx"
)

type SessionProc struct {
	Name  string  `json:"name"`
	CPU   float64 `json:"cpu"`
	GPU   float64 `json:"gpu"`
	RAMMB float64 `json:"ramMb"`
}

type SessionAction struct {
	ID         string            `json:"id"`
	Title      string            `json:"title"`
	What       string            `json:"what"`
	Tested     bool              `json:"tested"`
	Enabled    bool              `json:"enabled"`
	Status     string            `json:"status"`
	Detail     string            `json:"detail"`
	Procs      []SessionProc     `json:"procs"`
	VerifyID   string            `json:"verifyId"`
	Experiment *ExperimentRecord `json:"experiment"`
}

type SessionNote struct {
	ID     string        `json:"id"`
	Tone   string        `json:"tone"`
	Title  string        `json:"title"`
	Detail string        `json:"detail"`
	Procs  []SessionProc `json:"procs"`
}

type PrepCheck struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Detail  string `json:"detail"`
	Done    bool   `json:"done"`
	Action  string `json:"action"`
	Blocked string `json:"blocked"`
}

type SessionInfo struct {
	Enabled bool            `json:"enabled"`
	State   string          `json:"state"`
	Since   time.Time       `json:"since"`
	Actions []SessionAction `json:"actions"`
	Notes   []SessionNote   `json:"notes"`
	Prep    []PrepCheck     `json:"prep"`
}

type sessionConfig struct {
	Enabled    bool `json:"enabled"`
	Self       bool `json:"self"`
	Background bool `json:"background"`
	Windows    bool `json:"windows"`
	Power      bool `json:"power"`
	Watch      bool `json:"watch"`
}

type selfEntry struct {
	created int64
	prev    winx.Eco
}

type session struct {
	mu         sync.Mutex
	closed     bool
	cfg        sessionConfig
	state      string
	since      time.Time
	paused     int
	sampler    sysmon.CPUSampler
	gpu        *sysmon.GPUSampler
	gpuUse     map[uint32]float64
	procs      []sysmon.ProcCPU
	watch      watchLog
	self       map[uint32]selfEntry
	excused    map[string]bool
	launchedAt time.Time
	retryAt    time.Time
	onEnd      func(SessionNote)
	prep       []PrepCheck
	experiment map[string]*ExperimentRecord
}

const (
	stateOff     = "off"
	stateWaiting = "waiting"
	stateActive  = "active"
	statePaused  = "paused"
	launchGrace  = 2 * time.Minute
	restoreRetry = 30 * time.Second
)

var (
	throttlerOnce sync.Once
	throttlerInst *eco.Throttler
)

func throttler() *eco.Throttler {
	throttlerOnce.Do(func() { throttlerInst = eco.New(filepath.Join(winx.DataDir(), "eco.json")) })
	return throttlerInst
}

func sessionPath() string { return filepath.Join(winx.DataDir(), "session.json") }

func loadSession() *session {
	cfg := sessionConfig{Enabled: true, Self: true, Watch: true}
	if b, err := os.ReadFile(sessionPath()); err == nil {
		json.Unmarshal(b, &cfg)
	}
	return &session{cfg: cfg, state: stateOff, self: map[uint32]selfEntry{}, excused: map[string]bool{}}
}

func (ss *session) saveConfig() error {
	b, err := json.MarshalIndent(ss.cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sessionPath(), b, 0o644)
}

func (ss *session) want(running bool) string {
	switch {
	case ss.closed || !ss.cfg.Enabled:
		return stateOff
	case ss.paused > 0:
		return statePaused
	case running:
		return stateActive
	}
	return stateWaiting
}

func (ss *session) step(running, sample bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	next := ss.want(running)
	if ss.state == stateActive && next != stateActive {
		ss.stop()
		if n, ok := ss.summary(); ok && next == stateWaiting && ss.onEnd != nil {
			ss.onEnd(n)
		}
	}
	if ss.state != stateActive && next == stateActive {
		ss.start()
	}
	ss.state = next
	if next != stateActive {
		if next != statePaused && len(throttler().Throttled()) > 0 {
			throttler().RestoreAll()
		}
		if next != statePaused && time.Since(ss.launchedAt) > launchGrace && time.Now().After(ss.retryAt) {
			if restoreSystem() != nil {
				ss.retryAt = time.Now().Add(restoreRetry)
			}
		}
		return
	}
	if sample {
		ss.procs = ss.sampler.Sample()
		if ss.gpu != nil {
			ss.gpuUse = ss.gpu.Sample()
		}
	}
	ss.followForeground()
	if sample {
		ss.act()
	}
}

func (ss *session) start() {
	ss.since = time.Now()
	ss.watch = watchLog{}
	ss.excused = map[string]bool{}
	ss.sampler.Reset()
	ss.procs = ss.sampler.Sample()
	ss.gpu, _ = sysmon.NewGPUSampler()
	if ss.cfg.Windows {
		visualChange.on()
	}
	if ss.cfg.Power {
		powerChange.on()
	}
}

func (s *Service) prelaunch() {
	s.sess.mu.Lock()
	defer s.sess.mu.Unlock()
	if s.sess.cfg.Enabled && !s.sess.closed && s.sess.paused == 0 && s.sess.cfg.Windows {
		s.sess.launchedAt = time.Now()
		visualChange.on()
	}
}

func (ss *session) stop() {
	if ss.gpu != nil {
		ss.gpu.Close()
		ss.gpu = nil
	}
	ss.relaxSelf()
	throttler().RestoreAll()
	restoreSystem()
	ss.excused = map[string]bool{}
}

func (ss *session) summary() (SessionNote, bool) {
	if !ss.cfg.Watch {
		return SessionNote{}, false
	}
	for _, n := range ss.watch.notes() {
		if n.Tone == "warn" {
			return n, true
		}
	}
	return SessionNote{}, false
}

func (ss *session) act() {
	if ss.cfg.Watch {
		ss.watch.add(ss.procs, ss.ownTree(), ss.gpuUse)
	}
	if ss.cfg.Self && winx.ForegroundPID() != ownPID() {
		ss.calmSelf()
	}
	if ss.cfg.Background {
		audible, err := winx.AudiblePIDs()
		if err != nil {
			return
		}
		throttled := throttler().Throttled()
		for _, p := range ss.procs {
			if _, ok := throttled[p.Name]; ok && audible[p.PID] {
				throttler().RestoreExe(p.Name)
				ss.excused[p.Name] = true
			}
		}
		session, _ := winx.SessionOf(ownPID())
		throttler().Apply(eco.Pick(ss.procs, session, audible, ss.excused))
	}
}

func (ss *session) followForeground() {
	if winx.RobloxInFront() {
		clear(ss.excused)
		return
	}
	fg := winx.ForegroundPID()
	if fg == ownPID() {
		ss.relaxSelf()
		return
	}
	exe := ss.nameOf(fg)
	if exe == "" {
		return
	}
	if _, ok := throttler().Throttled()[exe]; ok {
		throttler().RestoreExe(exe)
	}
	if _, ok := eco.Label(exe); ok {
		ss.excused[exe] = true
	}
}

func (ss *session) nameOf(pid uint32) string {
	for _, p := range ss.procs {
		if p.PID == pid {
			return p.Name
		}
	}
	return ""
}

func ownPID() uint32 { return uint32(os.Getpid()) }

var ownStart = sync.OnceValue(func() int64 {
	created, _, _ := winx.ProcTimes(ownPID())
	return created
})

func (ss *session) ownTree() map[uint32]bool {
	byPID := make(map[uint32]sysmon.ProcCPU, len(ss.procs))
	for _, p := range ss.procs {
		byPID[p.PID] = p
	}
	own := map[uint32]bool{ownPID(): true}
	for _, p := range ss.procs {
		cur := p
		for hop := 0; hop < 6 && cur.Name == "msedgewebview2.exe" && cur.Created >= ownStart(); hop++ {
			if cur.Parent == ownPID() {
				own[p.PID] = true
				break
			}
			next, ok := byPID[cur.Parent]
			if !ok {
				break
			}
			cur = next
		}
	}
	return own
}

func (ss *session) calmSelf() {
	for pid := range ss.ownTree() {
		if pid == ownPID() {
			continue
		}
		if _, done := ss.self[pid]; done {
			continue
		}
		created, _, err := winx.ProcTimes(pid)
		if err != nil {
			continue
		}
		prev, err := winx.EcoOf(pid)
		if err != nil {
			continue
		}
		if winx.SetEco(pid, winx.Efficient()) == nil {
			ss.self[pid] = selfEntry{created: created, prev: prev}
		}
	}
}

func (ss *session) relaxSelf() {
	for pid, e := range ss.self {
		if c, _, err := winx.ProcTimes(pid); err == nil && c == e.created {
			winx.SetEco(pid, e.prev.Restorable())
		}
		delete(ss.self, pid)
	}
}

func (ss *session) current() string {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.state
}

func (s *Service) pauseSession() {
	s.sess.mu.Lock()
	defer s.sess.mu.Unlock()
	s.sess.paused++
	if s.sess.state == stateActive {
		s.sess.stop()
	}
	s.sess.state = statePaused
}

func (s *Service) resumeSession() {
	s.sess.mu.Lock()
	defer s.sess.mu.Unlock()
	if s.sess.paused > 0 {
		s.sess.paused--
	}
}

func (s *Service) Shutdown() {
	s.sess.mu.Lock()
	defer s.sess.mu.Unlock()
	s.sess.closed = true
	s.sess.stop()
	s.sess.state = stateOff
}

func (s *Service) Session() SessionInfo {
	prep, exp := s.prepChecks(), latestExperiments(backgroundID, visualID, powerID)
	s.sess.mu.Lock()
	s.sess.prep, s.sess.experiment = prep, exp
	s.sess.mu.Unlock()
	return s.sessionInfo()
}

func (s *Service) SetSession(enabled bool) (SessionInfo, error) {
	s.sess.mu.Lock()
	s.sess.cfg.Enabled = enabled
	err := s.sess.saveConfig()
	s.sess.mu.Unlock()
	s.sess.step(len(roblox.Running()) > 0, false)
	return s.sessionInfo(), err
}

func (s *Service) SetSessionAction(id string, on bool) (SessionInfo, error) {
	s.sess.mu.Lock()
	switch id {
	case "self":
		s.sess.cfg.Self = on
		if !on {
			s.sess.relaxSelf()
		}
	case "background":
		s.sess.cfg.Background = on
		if !on {
			throttler().RestoreAll()
		}
	case "windows":
		s.sess.cfg.Windows = on
		switch {
		case !on:
			visualChange.off()
		case s.sess.state == stateActive:
			visualChange.on()
		}
	case "power":
		s.sess.cfg.Power = on
		switch {
		case !on:
			powerChange.off()
		case s.sess.state == stateActive:
			powerChange.on()
		}
	case "watch":
		s.sess.cfg.Watch = on
	default:
		s.sess.mu.Unlock()
		return s.sessionInfo(), errors.New(i18n.T("acción desconocida"))
	}
	err := s.sess.saveConfig()
	s.sess.mu.Unlock()
	return s.sessionInfo(), err
}
