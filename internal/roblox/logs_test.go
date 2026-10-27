package roblox

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseLineReadsServerStart(t *testing.T) {
	line := "2026-10-27T15:46:42.196Z,8.196402,0768,6 [FLog::Output] Server Prefix: 0.745.0.7451137_20261027T154617Z_RCC_21fe0"
	ev, ok := ParseLine(line)
	want := time.Date(2026, 10, 27, 15, 46, 17, 0, time.UTC)
	if !ok || ev.Kind != EventServerStart || !ev.Started.Equal(want) {
		t.Fatalf("ParseLine = %+v, %v; want server start %v", ev, ok, want)
	}
}

func TestServerStartFindsPrefixInLog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "player.log")
	lines := "2026-10-27T15:46:38.769Z,4.769841,269c,6 [FLog::Output] ! Joining game '3ea23306-7b19-4f9a-865b-2cad443ba709' place 14315258385 at 10.204.0.29\n" +
		"2026-10-27T15:46:42.196Z,8.196402,0768,6 [FLog::Output] Server Prefix: 0.745.0.7451137_20261027T154617Z_RCC_21fe0\n"
	if err := os.WriteFile(path, []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	begin := time.Now()
	started, ok := ServerStart(context.Background(), Session{LogPath: path}, 5*time.Second)
	if !ok || !started.Equal(time.Date(2026, 10, 27, 15, 46, 17, 0, time.UTC)) {
		t.Fatalf("ServerStart = %v, %v", started, ok)
	}
	if waited := time.Since(begin); waited > 2*time.Second {
		t.Fatalf("ServerStart waited %v after finding the prefix", waited)
	}
	if _, ok := ServerStart(context.Background(), Session{LogPath: filepath.Join(t.TempDir(), "missing.log")}, 300*time.Millisecond); ok {
		t.Fatal("ServerStart found a start time in a missing log")
	}
}

func TestParseLineReadsDatacenter(t *testing.T) {
	line := "2026-10-27T15:21:52.973Z,2.973706,4a04,6,Info [DFLog::NetworkClient] Transport selection: useRbxTransportEnabled=false, selectedTransport=RakNet, rccPort=58698, DatacenterId=25774"
	ev, ok := ParseLine(line)
	if !ok || ev.Kind != EventDatacenter || ev.Datacenter != 25774 {
		t.Fatalf("ParseLine = %+v, %v; want datacenter 25774", ev, ok)
	}
}
