package app

import (
	"context"
	"sync"

	"toque/internal/bench"
	"toque/internal/catalog"
	"toque/internal/roblox"
	"toque/internal/session"
)

type liveSession struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	done    chan struct{}
	release func()
}

func (s *Service) Recover() {
	if inst, err := roblox.Find(); err == nil {
		bench.RecoverInterrupted(inst)
	}
}

func (s *Service) StartSession() {
	s.live.mu.Lock()
	defer s.live.mu.Unlock()
	if s.live.cancel != nil {
		return
	}
	ok, release := session.Exclusive(`Local\ToqueSession`)
	if !ok {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	ctl := session.NewController(catalog.LiveProven())
	s.live.cancel, s.live.done, s.live.release = cancel, done, release
	go func() {
		ctl.Run(ctx)
		close(done)
	}()
}

func (s *Service) StopSession() {
	s.live.mu.Lock()
	cancel, done, release := s.live.cancel, s.live.done, s.live.release
	s.live.cancel, s.live.done, s.live.release = nil, nil, nil
	s.live.mu.Unlock()
	if cancel == nil {
		return
	}
	cancel()
	<-done
	release()
}

func (s *Service) pauseLive() func() {
	agentWasRunning := AgentRunning()
	if agentWasRunning {
		stopAgent()
	}
	s.StopSession()
	return func() {
		if agentWasRunning {
			startAgent()
			return
		}
		s.StartSession()
	}
}
