package roblox

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"
)

type LogState struct {
	JobID      string
	PlaceID    int64
	ServerIP   string
	ServerPort int
}

type Tail struct {
	mu     sync.Mutex
	path   string
	offset int64
	state  LogState
	rest   string
}

func (t *Tail) Read(path string) LogState {
	t.mu.Lock()
	defer t.mu.Unlock()
	if path != t.path {
		t.path, t.offset, t.state, t.rest = path, 0, LogState{}, ""
	}
	f, err := os.Open(path)
	if err != nil {
		return t.state
	}
	defer f.Close()
	if info, err := f.Stat(); err == nil && info.Size() < t.offset {
		t.offset, t.state, t.rest = 0, LogState{}, ""
	}
	if _, err := f.Seek(t.offset, io.SeekStart); err != nil {
		return t.state
	}
	r := bufio.NewReaderSize(f, 64*1024)
	for {
		chunk, err := r.ReadString('\n')
		t.offset += int64(len(chunk))
		if strings.HasSuffix(chunk, "\n") {
			line := strings.TrimRight(t.rest+chunk, "\r\n")
			t.rest = ""
			if ev, ok := ParseLine(line); ok {
				t.apply(ev)
			}
		} else {
			t.rest += chunk
		}
		if err != nil {
			break
		}
	}
	return t.state
}

func (t *Tail) apply(ev Event) {
	switch ev.Kind {
	case EventJoining:
		t.state = LogState{JobID: ev.JobID, PlaceID: ev.PlaceID}
	case EventUdmux:
		t.state.ServerIP, t.state.ServerPort = ev.IP, ev.Port
	case EventConnected:
		if t.state.ServerIP == "" {
			t.state.ServerIP, t.state.ServerPort = ev.IP, ev.Port
		}
	case EventDisconnected:
		t.state.ServerIP = ""
	}
}
