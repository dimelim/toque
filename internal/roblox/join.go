package roblox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"toque/internal/winx"
)

const RealisticStreetSoccer int64 = 14315258385

type Session struct {
	PID        uint32
	LogPath    string
	JobID      string
	PlaceID    int64
	ServerIP   string
	ServerPort int
	JoinSecs   float64
	LaunchedAt time.Time
	JoinedAt   time.Time
}

func (s Session) TimeToJoin() time.Duration {
	if s.JoinedAt.IsZero() {
		return 0
	}
	return s.JoinedAt.Sub(s.LaunchedAt)
}

func JoinURL(placeID int64, jobID string) string {
	u := fmt.Sprintf("roblox://experiences/start?placeId=%d", placeID)
	if jobID != "" {
		u += "&gameInstanceId=" + jobID
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

func Join(ctx context.Context, inst Install, placeID int64, jobID string, timeout time.Duration) (Session, error) {
	CloseAll(10 * time.Second)
	time.Sleep(1500 * time.Millisecond)
	s := Session{PlaceID: placeID, LaunchedAt: time.Now()}
	pid, err := winx.LaunchAsUser(inst.Player, JoinURL(placeID, jobID), inst.VersionDir)
	if err != nil {
		return s, err
	}
	s.PID = pid
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
	for {
		select {
		case ev := <-events:
			switch ev.Kind {
			case EventJoining:
				s.JobID, s.PlaceID = ev.JobID, ev.PlaceID
			case EventUdmux:
				s.ServerIP, s.ServerPort = ev.IP, ev.Port
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
		case <-ctx.Done():
			return s, errors.New("join not detected before timeout")
		}
	}
}
