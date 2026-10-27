package roblox

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type EventKind int

const (
	EventJoining EventKind = iota + 1
	EventUdmux
	EventLoadTime
	EventConnected
	EventDisconnected
)

type Event struct {
	Time    time.Time
	Kind    EventKind
	JobID   string
	PlaceID int64
	IP      string
	Port    int
	Seconds float64
}

var (
	reJoining  = regexp.MustCompile(`! Joining game '([0-9a-f-]{36})' place (\d+) at ([0-9.]+)`)
	reUdmux    = regexp.MustCompile(`UDMUX Address = ([0-9.]+), Port = (\d+)`)
	reLoadTime = regexp.MustCompile(`game_join_loadtime: placeid:(\d+), join_time:([0-9.]+)`)
)

func ParseLine(line string) (Event, bool) {
	ts := time.Time{}
	if len(line) > 24 {
		if t, err := time.Parse("2006-01-02T15:04:05.000Z", line[:24]); err == nil {
			ts = t
		}
	}
	switch {
	case strings.Contains(line, "! Joining game"):
		if m := reJoining.FindStringSubmatch(line); m != nil {
			id, _ := strconv.ParseInt(m[2], 10, 64)
			return Event{Time: ts, Kind: EventJoining, JobID: m[1], PlaceID: id, IP: m[3]}, true
		}
	case strings.Contains(line, "UDMUX Address"):
		if m := reUdmux.FindStringSubmatch(line); m != nil {
			port, _ := strconv.Atoi(m[2])
			return Event{Time: ts, Kind: EventUdmux, IP: m[1], Port: port}, true
		}
	case strings.Contains(line, "game_join_loadtime"):
		if m := reLoadTime.FindStringSubmatch(line); m != nil {
			id, _ := strconv.ParseInt(m[1], 10, 64)
			s, _ := strconv.ParseFloat(m[2], 64)
			return Event{Time: ts, Kind: EventLoadTime, PlaceID: id, Seconds: s}, true
		}
	case strings.Contains(line, "Replicator created for player"):
		return Event{Time: ts, Kind: EventConnected}, true
	case strings.Contains(line, "Client:Disconnect"):
		return Event{Time: ts, Kind: EventDisconnected}, true
	}
	return Event{}, false
}

func NewestPlayerLog(dir string, since time.Time) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	var best string
	var bestTime time.Time
	for _, e := range entries {
		if e.IsDir() || !IsPlayerLog(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(since) {
			continue
		}
		if best == "" || info.ModTime().After(bestTime) {
			best, bestTime = filepath.Join(dir, e.Name()), info.ModTime()
		}
	}
	return best
}

func Follow(ctx context.Context, path string, fromStart bool, emit func(Event)) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if !fromStart {
		f.Seek(0, io.SeekEnd)
	}
	r := bufio.NewReaderSize(f, 64*1024)
	var partial strings.Builder
	for {
		line, err := r.ReadString('\n')
		if len(line) > 0 {
			partial.WriteString(line)
			if strings.HasSuffix(line, "\n") {
				if ev, ok := ParseLine(strings.TrimRight(partial.String(), "\r\n")); ok {
					emit(ev)
				}
				partial.Reset()
			}
		}
		if err == io.EOF {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
			}
			continue
		}
		if err != nil {
			return err
		}
	}
}

func ScanFile(path string) []Event {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 256*1024), 4*1024*1024)
	for sc.Scan() {
		if ev, ok := ParseLine(sc.Text()); ok {
			out = append(out, ev)
		}
	}
	return out
}
