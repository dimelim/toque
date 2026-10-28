package roblox

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/sys/windows/registry"

	"toque/internal/winx"
)

const PlayerExe = "RobloxPlayerBeta.exe"

type Install struct {
	Root       string
	VersionDir string
	Player     string
	Version    string
}

var (
	commandExe = regexp.MustCompile(`(?i)^\s*"([^"]+\.exe)"|^\s*(\S+\.exe)`)
	lastPlayer atomic.Value
)

func Find() (Install, error) {
	local := winx.UserLocalAppData()
	root := filepath.Join(local, "Roblox")
	player := protocolPlayer()
	if player == "" {
		player = newestPlayer(versionDirs(local, root)...)
	}
	if player == "" {
		player = runningPlayer()
	}
	if player == "" {
		return Install{Root: root}, errors.New("roblox player not installed")
	}
	dir := filepath.Dir(player)
	return Install{Root: root, VersionDir: dir, Player: player, Version: filepath.Base(dir)}, nil
}

func versionDirs(local, root string) []string {
	dirs := []string{
		filepath.Join(root, "Versions"),
		filepath.Join(local, "Bloxstrap", "Versions"),
		filepath.Join(local, "Fishstrap", "Versions"),
		filepath.Join(local, "Voidstrap", "RblxVersions"),
	}
	for _, env := range []string{"ProgramFiles(x86)", "ProgramFiles"} {
		if pf := os.Getenv(env); pf != "" {
			dirs = append(dirs, filepath.Join(pf, "Roblox", "Versions"))
		}
	}
	return dirs
}

func newestPlayer(dirs ...string) string {
	var best string
	var bestTime time.Time
	for _, dir := range dirs {
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			p := filepath.Join(dir, e.Name(), PlayerExe)
			if info, err := os.Stat(p); e.IsDir() && err == nil && !info.IsDir() && info.ModTime().After(bestTime) {
				best, bestTime = p, info.ModTime()
			}
		}
	}
	return best
}

func exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func protocolPlayer() string {
	for _, exe := range protocolTargets() {
		if !filepath.IsAbs(exe) {
			continue
		}
		dir := filepath.Dir(exe)
		if p := filepath.Join(dir, PlayerExe); exists(p) {
			return p
		}
		if p := newestPlayer(filepath.Join(dir, "Versions"), filepath.Join(dir, "RblxVersions")); p != "" {
			return p
		}
	}
	return ""
}

func protocolTargets() []string {
	var roots []registry.Key
	if user, err := winx.UserRoot(); err == nil {
		roots = append(roots, user)
	}
	roots = append(roots, registry.LOCAL_MACHINE)
	var out []string
	for _, root := range roots {
		for _, scheme := range []string{"roblox-player", "roblox"} {
			k, err := registry.OpenKey(root, `Software\Classes\`+scheme+`\shell\open\command`, registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			v, _, err := k.GetStringValue("")
			k.Close()
			if exe := commandTarget(v); err == nil && exe != "" {
				out = append(out, exe)
			}
		}
	}
	return out
}

func commandTarget(command string) string {
	m := commandExe.FindStringSubmatch(command)
	if m == nil {
		return ""
	}
	if m[1] != "" {
		return m[1]
	}
	return m[2]
}

func runningPlayer() string {
	for _, p := range winx.FindByName(PlayerExe) {
		if path, err := winx.ImagePath(p.PID); err == nil && exists(path) {
			lastPlayer.Store(path)
			return path
		}
	}
	if p, ok := lastPlayer.Load().(string); ok && exists(p) {
		return p
	}
	return ""
}

func (i Install) LogDir() string {
	return filepath.Join(i.Root, "logs")
}

func (i Install) SettingsPath() string {
	return filepath.Join(i.Root, "GlobalBasicSettings_13.xml")
}

func Running() []uint32 {
	var out []uint32
	for _, p := range winx.FindByName(PlayerExe) {
		out = append(out, p.PID)
	}
	return out
}

func IsPlayerLog(name string) bool {
	return strings.Contains(name, "_Player_") && !strings.Contains(name, "CrashHandler") && strings.HasSuffix(name, ".log")
}
