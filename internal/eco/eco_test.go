package eco

import (
	"os/exec"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"

	"toque/internal/sysmon"
	"toque/internal/winx"
)

func startTarget(t *testing.T) Target {
	cmd := exec.Command("ping", "-n", "30", "127.0.0.1")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cmd.Process.Kill(); cmd.Wait() })
	pid := uint32(cmd.Process.Pid)
	created, _, err := winx.ProcTimes(pid)
	if err != nil {
		t.Fatal(err)
	}
	return Target{PID: pid, Created: created, Exe: "ping.exe"}
}

func classOf(t *testing.T, pid uint32) uint32 {
	e, err := winx.EcoOf(pid)
	if err != nil {
		t.Fatal(err)
	}
	return e.Class
}

func TestRecoverRestoresAfterCrash(t *testing.T) {
	path := filepath.Join(t.TempDir(), "eco.json")
	tg := startTarget(t)
	before := classOf(t, tg.PID)
	New(path).Apply([]Target{tg})
	if c := classOf(t, tg.PID); c != windows.IDLE_PRIORITY_CLASS {
		t.Fatalf("class after apply = %#x", c)
	}
	if err := New(path).Recover(); err != nil {
		t.Fatal(err)
	}
	if c := classOf(t, tg.PID); c != before {
		t.Fatalf("class after recover = %#x, want %#x", c, before)
	}
}

func TestRecoverSkipsReusedPID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "eco.json")
	tg := startTarget(t)
	th := New(path)
	th.Apply([]Target{tg})
	th.entries[tg.PID] = entry{PID: tg.PID, Created: tg.Created + 1, Exe: tg.Exe, Prev: winx.Eco{Class: windows.HIGH_PRIORITY_CLASS}}
	th.save()
	if err := New(path).Recover(); err != nil {
		t.Fatal(err)
	}
	if c := classOf(t, tg.PID); c != windows.IDLE_PRIORITY_CLASS {
		t.Fatalf("a different process was touched, class = %#x", c)
	}
}

func TestPickSkipsQuietLoudAndExcusedApps(t *testing.T) {
	procs := []sysmon.ProcCPU{
		{PID: 1, Name: "chrome.exe", Core: 4, Session: 1},
		{PID: 2, Name: "chrome.exe", Core: 3, Session: 1},
		{PID: 3, Name: "steam.exe", Core: 9, Session: 1},
		{PID: 4, Name: "steamwebhelper.exe", Core: 20, Session: 1},
		{PID: 5, Name: "onedrive.exe", Core: 30, Session: 1},
		{PID: 6, Name: "discord.exe", Core: 50, Session: 1},
		{PID: 7, Name: "dropbox.exe", Core: 30, Session: 0},
		{PID: 8, Name: "brave.exe", Core: 2, Session: 1},
	}
	got := Pick(procs, 1, map[uint32]bool{4: true}, map[string]bool{"onedrive.exe": true})
	want := map[uint32]bool{1: true, 2: true, 3: true}
	if len(got) != len(want) {
		t.Fatalf("Pick = %+v", got)
	}
	for _, g := range got {
		if !want[g.PID] {
			t.Fatalf("Pick picked %+v", g)
		}
	}
}
