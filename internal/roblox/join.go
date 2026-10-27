package roblox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"toque/internal/winx"
)

const RealisticStreetSoccer int64 = 14315258385

type Session struct {
	PID          uint32
	LogPath      string
	JobID        string
	PlaceID      int64
	ServerIP     string
	ServerPort   int
	DatacenterID int
	JoinSecs     float64
	LaunchedAt   time.Time
	JoinedAt     time.Time
}

func (s Session) TimeToJoin() time.Duration {
	if s.JoinedAt.IsZero() {
		return 0
	}
	return s.JoinedAt.Sub(s.LaunchedAt)
}

type Target struct {
	PlaceID   int64
	JobID     string
	LinkCode  string
	ShareCode string
}

func (t Target) URL() string {
	if t.ShareCode != "" {
		return "roblox://navigation/share_links?code=" + url.QueryEscape(t.ShareCode) + "&type=Server"
	}
	u := fmt.Sprintf("roblox://experiences/start?placeId=%d", t.PlaceID)
	if t.JobID != "" {
		u += "&gameInstanceId=" + url.QueryEscape(t.JobID)
	}
	if t.LinkCode != "" {
		u += "&linkCode=" + url.QueryEscape(t.LinkCode)
	}
	return u
}

func CloseAll(timeout time.Duration) {
	for _, pid := range Running() {
		winx.Terminate(pid)
	}
	deadline := time.Now().Add(timeout)
	for len(Running()) > 0 && time.Now().Before(deadline) {
		time.Sleep(200 * time.Millisecond)
	}
}

var (
	ErrRejected = errors.New("server rejected")
	ErrQueued   = errors.New("roblox put the join in a queue")
)

const queueWait = 15 * time.Second

func Join(ctx context.Context, inst Install, placeID int64, jobID string, timeout time.Duration) (Session, error) {
	return JoinIf(ctx, inst, Target{PlaceID: placeID, JobID: jobID}, timeout, nil)
}

func JoinIf(ctx context.Context, inst Install, t Target, timeout time.Duration, accept func(Session) bool) (Session, error) {
	CloseAll(10 * time.Second)
	s := Session{PlaceID: t.PlaceID}
	select {
	case <-ctx.Done():
		return s, ctx.Err()
	case <-time.After(1500 * time.Millisecond):
	}
	s.LaunchedAt = time.Now()
	pid, err := winx.LaunchAsUser(inst.Player, t.URL(), inst.VersionDir)
	if err != nil {
		return s, err
	}
	s.PID = pid
	s, err = watchJoin(ctx, inst, s, t.JobID != "", timeout, accept)
	if err != nil {
		CloseAll(10 * time.Second)
		if ctx.Err() != nil {
			return s, ctx.Err()
		}
	}
	return s, err
}

func watchJoin(ctx context.Context, inst Install, s Session, canQueue bool, timeout time.Duration, accept func(Session) bool) (Session, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for s.LogPath == "" {
		if s.LogPath = NewestPlayerLog(inst.LogDir(), s.LaunchedAt.Add(-time.Second)); s.LogPath != "" {
			break
		}
		select {
		case <-ctx.Done():
			return s, errors.New("roblox did not create a log")
		case <-time.After(250 * time.Millisecond):
		}
	}
	events := make(chan Event, 64)
	go func() {
		Follow(ctx, s.LogPath, true, func(ev Event) {
			select {
			case events <- ev:
			case <-ctx.Done():
			}
		})
	}()
	var queued <-chan time.Time
	if canQueue {
		timer := time.NewTimer(queueWait)
		defer timer.Stop()
		queued = timer.C
	}
	for {
		select {
		case ev := <-events:
			switch ev.Kind {
			case EventJoining:
				s.JobID, s.PlaceID = ev.JobID, ev.PlaceID
				queued = nil
			case EventDatacenter:
				s.DatacenterID = ev.Datacenter
			case EventUdmux:
				s.ServerIP, s.ServerPort = ev.IP, ev.Port
				if accept != nil && !accept(s) {
					return s, ErrRejected
				}
			case EventLoadTime:
				s.JoinSecs = ev.Seconds
			case EventConnected:
				if s.ServerIP != "" {
					s.JoinedAt = time.Now()
					if pids := Running(); len(pids) > 0 {
						s.PID = pids[0]
					}
					return s, nil
				}
			}
		case <-queued:
			return s, ErrQueued
		case <-ctx.Done():
			return s, errors.New("join not detected before timeout")
		}
	}
}
