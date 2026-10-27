package roblox

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchJoinFollowsRelaunchedLogWithoutUdmux(t *testing.T) {
	root := t.TempDir()
	logs := filepath.Join(root, "logs")
	if err := os.MkdirAll(logs, 0o755); err != nil {
		t.Fatal(err)
	}
	first := filepath.Join(logs, "0.745_20261027T220000Z_Player_AAAAA_last.log")
	if err := os.WriteFile(first, []byte("starting\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(1500 * time.Millisecond)
		lines := "2026-10-27T16:42:36.696Z,4.696810,2b8c,6 [FLog::Output] ! Joining game 'a36d3c5b-198f-403b-8c38-a63f0cd08632' place 14315258385 at 10.8.3.223\n" +
			"2026-10-27T16:42:37.376Z,5.376985,0a28,7 [FLog::Network] Replicator created for player 128.116.95.33|55422\n"
		os.WriteFile(filepath.Join(logs, "0.745_20261027T220005Z_Player_BBBBB_last.log"), []byte(lines), 0o644)
	}()
	s, err := watchJoin(context.Background(), Install{Root: root}, Session{LaunchedAt: time.Now()}, false, 10*time.Second, nil)
	if err != nil {
		t.Fatalf("watchJoin: %v", err)
	}
	if s.ServerIP != "128.116.95.33" || s.ServerPort != 55422 || s.JobID != "a36d3c5b-198f-403b-8c38-a63f0cd08632" {
		t.Fatalf("watchJoin = %+v", s)
	}
}
