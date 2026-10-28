package app

import (
	"bytes"
	"encoding/json"
	"time"

	"toque/internal/winx"
)

type WindowState struct {
	Calm bool `json:"calm"`
}

const (
	overviewEvery = 3
	sampleEvery   = 10
)

func (s *Service) send(name string, data any) {
	if s.emit != nil {
		s.emit(name, data)
	}
}

func sendIfChanged(s *Service, name string, data any, last *[]byte) {
	b, err := json.Marshal(data)
	if err != nil || bytes.Equal(b, *last) {
		return
	}
	*last = b
	s.send(name, data)
}

func (s *Service) StartMonitor(minimized func() bool) {
	go s.monitor(minimized)
}

func (s *Service) Window() WindowState {
	return WindowState{Calm: s.calm.Load()}
}

func (s *Service) monitor(minimized func() bool) {
	var lastOverview, lastSession []byte
	var o Overview
	calm := false
	for tick := 0; ; tick++ {
		now := minimized() || winx.RobloxInFront()
		changed := tick == 0 || now != calm
		if changed {
			calm = now
			s.calm.Store(calm)
			s.send("window", WindowState{Calm: calm})
			lastOverview, lastSession = nil, nil
		}
		if changed || tick%overviewEvery == 0 {
			if calm {
				o.Roblox = robloxState()
				s.feedPing(o)
			} else {
				o = s.Overview()
			}
		}
		s.sess.step(o.Roblox.Running, tick%sampleEvery == 0)
		if !calm {
			sendIfChanged(s, "overview", o, &lastOverview)
			sendIfChanged(s, "session", s.sessionInfo(), &lastSession)
		}
		time.Sleep(time.Second)
	}
}
