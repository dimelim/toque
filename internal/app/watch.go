package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"toque/internal/roblox"
)

type WatchInfo struct {
	Region  string       `json:"region"`
	Active  bool         `json:"active"`
	Since   time.Time    `json:"since"`
	Checked int          `json:"checked"`
	Joined  *ServerProbe `json:"joined,omitempty"`
	Note    string       `json:"note,omitempty"`
}

const (
	watchEvery = 25 * time.Second
	watchFor   = time.Hour
)

func (s *Service) WatchRegion(key string) WatchInfo {
	ctx, cancel := context.WithTimeout(context.Background(), watchFor)
	s.watchMu.Lock()
	if s.watchStop != nil {
		s.watchStop()
	}
	s.watchGen++
	gen := s.watchGen
	s.watchStop = cancel
	s.watch = WatchInfo{Region: key, Active: true, Since: time.Now()}
	info := s.watch
	s.watchMu.Unlock()
	s.emitWatch(info)
	go s.watchLoop(ctx, gen, key)
	return info
}

func (s *Service) StopWatch() WatchInfo {
	s.watchMu.Lock()
	if s.watchStop != nil {
		s.watchStop()
		s.watchStop = nil
	}
	s.watchGen++
	s.watch.Active = false
	info := s.watch
	s.watchMu.Unlock()
	s.emitWatch(info)
	return info
}

func (s *Service) WatchStatus() WatchInfo {
	s.watchMu.Lock()
	defer s.watchMu.Unlock()
	return s.watch
}

func (s *Service) emitWatch(info WatchInfo) {
	if s.emit != nil {
		s.emit("watch", info)
	}
}

func (s *Service) updateWatch(gen int, fn func(*WatchInfo)) {
	s.watchMu.Lock()
	if gen != s.watchGen {
		s.watchMu.Unlock()
		return
	}
	fn(&s.watch)
	info := s.watch
	s.watchMu.Unlock()
	s.emitWatch(info)
}

func (s *Service) endWatch(gen int) {
	s.watchMu.Lock()
	defer s.watchMu.Unlock()
	if gen == s.watchGen && s.watchStop != nil {
		s.watchStop()
		s.watchStop = nil
	}
}

func (s *Service) watchLoop(ctx context.Context, gen int, key string) {
	defer s.endWatch(gen)
	defer s.updateWatch(gen, func(w *WatchInfo) { w.Active = false })
	cursor := ""
	for {
		wait := watchEvery
		if s.idle() == nil {
			page, next, err := roblox.PublicPage(ctx, roblox.RealisticStreetSoccer, roblox.EmptiestFirst, cursor)
			var limited *roblox.RateLimitError
			switch {
			case errors.As(err, &limited):
				wait = limited.Wait
			case err == nil:
				cursor = next
				s.updateWatch(gen, func(w *WatchInfo) { w.Checked += len(page) })
				if matches, _ := serversIn(ctx, key, page); len(matches) > 0 {
					playing := len(roblox.Running()) > 0
					if probe, ok := s.joinWatched(key, matches); ok {
						s.updateWatch(gen, func(w *WatchInfo) { w.Joined = &probe })
						return
					}
					if playing && ctx.Err() == nil {
						s.updateWatch(gen, func(w *WatchInfo) {
							w.Note = fmt.Sprintf("No pude entrar en la partida de %s, así que paré el vigía para no volver a sacarte de tu partida.", regionCity(key))
						})
						return
					}
				}
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}

func (s *Service) joinWatched(key string, matches []roblox.PublicServer) (ServerProbe, bool) {
	ctx, done, err := s.begin("play")
	if err != nil {
		return ServerProbe{}, false
	}
	defer done()
	inst, err := s.launchInstall()
	if err != nil {
		return ServerProbe{}, false
	}
	s.progress("play", "join", 0.1, fmt.Sprintf("Hay una partida en %s. Entrando", regionCity(key)))
	sess, _, err := s.joinIn(ctx, inst, key, matches, "play")
	if err != nil {
		return ServerProbe{}, false
	}
	return joinedProbe(ctx, sess), true
}
