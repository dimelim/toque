package app

import (
	"runtime"
	"testing"

	"toque/internal/sysmon"
)

func TestFramePresetPicksOfficialValueAtLeastRefresh(t *testing.T) {
	cases := map[int]int{60: 60, 75: 120, 100: 120, 120: 120, 144: 144, 165: 240, 240: 240, 360: 240}
	for hz, want := range cases {
		if got := framePreset(hz); got != want {
			t.Errorf("framePreset(%d) = %d, want %d", hz, got, want)
		}
	}
}

func TestWatchNotesGroupAppsAndHideToque(t *testing.T) {
	cores := float64(runtime.NumCPU())
	var w watchLog
	own := map[uint32]bool{10: true, 11: true}
	procs := []sysmon.ProcCPU{
		{PID: 10, Name: "toque.exe", Core: 50 * cores},
		{PID: 11, Name: "msedgewebview2.exe", Core: 50 * cores},
		{PID: 20, Name: "steam.exe", Core: 3 * cores},
		{PID: 21, Name: "steamwebhelper.exe", Core: 4 * cores},
		{PID: 30, Name: "msmpeng.exe", Core: 6 * cores},
		{PID: 40, Name: "exitlag.exe", Core: 0},
		{PID: 50, Name: "svchost.exe", Core: 20 * cores},
		{PID: 60, Name: "robloxplayerbeta.exe", Core: 30 * cores},
	}
	gpu := map[uint32]float64{20: 12, 60: 80, 11: 40}
	w.add(procs, own, gpu)
	w.add(procs, own, gpu)
	ids := map[string]SessionNote{}
	for _, n := range w.notes() {
		ids[n.ID] = n
	}
	cpu, ok := ids["cpu"]
	if !ok || len(cpu.Procs) != 1 || cpu.Procs[0].Name != "Steam" || cpu.Procs[0].CPU != 7 {
		t.Fatalf("cpu note = %+v", cpu)
	}
	if cpu.Tone != "warn" {
		t.Errorf("7 %% of the CPU should warn, got %q", cpu.Tone)
	}
	if g, ok := ids["gpu"]; !ok || len(g.Procs) != 1 || g.Procs[0].Name != "Steam" || g.Procs[0].GPU != 12 || g.Tone != "warn" {
		t.Fatalf("gpu note = %+v", g)
	}
	if _, ok := ids["defender"]; !ok {
		t.Error("missing defender note")
	}
	if _, ok := ids["exitlag"]; !ok {
		t.Error("missing exitlag note")
	}
	if got := share(w.own, w.samples); got != 100 {
		t.Errorf("own share = %v, want 100", got)
	}
}

func TestSessionStateFollowsSwitchPauseAndRoblox(t *testing.T) {
	ss := &session{cfg: sessionConfig{Enabled: true}, state: stateOff, self: map[uint32]selfEntry{}, excused: map[string]bool{}}
	if got := ss.want(false); got != stateWaiting {
		t.Fatalf("enabled without roblox = %q", got)
	}
	if got := ss.want(true); got != stateActive {
		t.Fatalf("enabled with roblox = %q", got)
	}
	ss.paused = 1
	if got := ss.want(true); got != statePaused {
		t.Fatalf("paused with roblox = %q", got)
	}
	ss.cfg.Enabled = false
	if got := ss.want(true); got != stateOff {
		t.Fatalf("disabled = %q", got)
	}
	ss.cfg.Enabled, ss.paused, ss.closed = true, 0, true
	if got := ss.want(true); got != stateOff {
		t.Fatalf("after shutdown = %q", got)
	}
}

func TestOwnTreeKeepsToqueWebViewAndNeverRoblox(t *testing.T) {
	start := ownStart()
	me := ownPID()
	ss := &session{procs: []sysmon.ProcCPU{
		{PID: me, Name: "toque.exe", Created: start},
		{PID: 101, Parent: me, Name: "msedgewebview2.exe", Created: start + 1},
		{PID: 102, Parent: 101, Name: "msedgewebview2.exe", Created: start + 2},
		{PID: 103, Parent: me, Name: "robloxplayerbeta.exe", Created: start + 3},
		{PID: 104, Parent: 103, Name: "robloxcrashhandler.exe", Created: start + 4},
		{PID: 105, Parent: me, Name: "msedgewebview2.exe", Created: start - 1},
		{PID: 106, Parent: 999, Name: "msedgewebview2.exe", Created: start + 5},
	}}
	own := ss.ownTree()
	for pid, want := range map[uint32]bool{me: true, 101: true, 102: true, 103: false, 104: false, 105: false, 106: false} {
		if own[pid] != want {
			t.Errorf("own[%d] = %v, want %v", pid, own[pid], want)
		}
	}
}
