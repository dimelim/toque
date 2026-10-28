package roblox

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCommandTarget(t *testing.T) {
	cases := map[string]string{
		`"C:\Users\Ana\AppData\Local\Roblox\Versions\version-1\RobloxPlayerBeta.exe" %1`: `C:\Users\Ana\AppData\Local\Roblox\Versions\version-1\RobloxPlayerBeta.exe`,
		`"C:\Users\Ana\AppData\Local\Bloxstrap\Bloxstrap.exe" -player "%1"`:              `C:\Users\Ana\AppData\Local\Bloxstrap\Bloxstrap.exe`,
		`C:\Users\Ana\AppData\Local\Fishstrap\Fishstrap.exe -player "%1"`:                `C:\Users\Ana\AppData\Local\Fishstrap\Fishstrap.exe`,
		`  "D:\Juegos\Roblox\Versions\version-2\RobloxPlayerBeta.EXE" %1`:                `D:\Juegos\Roblox\Versions\version-2\RobloxPlayerBeta.EXE`,
		`rundll32.dll,Something`: "",
	}
	for command, want := range cases {
		if got := commandTarget(command); got != want {
			t.Errorf("commandTarget(%q) = %q, want %q", command, got, want)
		}
	}
}

func TestNewestPlayerAcrossFolders(t *testing.T) {
	base := t.TempDir()
	player := func(parts ...string) string {
		dir := filepath.Join(append([]string{base}, parts...)...)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(dir, PlayerExe)
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	old := player("Roblox", "Versions", "version-old")
	fresh := player("Bloxstrap", "Versions", "version-new")
	now := time.Now()
	os.Chtimes(old, now.Add(-time.Hour), now.Add(-time.Hour))
	os.Chtimes(fresh, now, now)
	got := newestPlayer(filepath.Join(base, "Roblox", "Versions"), filepath.Join(base, "Bloxstrap", "Versions"), filepath.Join(base, "Missing"))
	if got != fresh {
		t.Fatalf("newestPlayer = %q, want %q", got, fresh)
	}
	bracketed := player("Juegos [PC]", "Versions", "version-x")
	if got := newestPlayer(filepath.Join(base, "Juegos [PC]", "Versions")); got != bracketed {
		t.Fatalf("brackets in the path should not break the search, got %q", got)
	}
}
