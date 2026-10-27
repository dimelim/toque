package roblox

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

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

var quotedExe = regexp.MustCompile(`"([^"]+RobloxPlayerBeta\.exe)"`)

func Find() (Install, error) {
	root := filepath.Join(winx.LocalAppData(), "Roblox")
	if p := playerFromProtocol(); p != "" {
		if _, err := os.Stat(p); err == nil {
			dir := filepath.Dir(p)
			if versions := filepath.Dir(dir); strings.EqualFold(filepath.Base(versions), "Versions") {
				root = filepath.Dir(versions)
			}
			return Install{Root: root, VersionDir: dir, Player: p, Version: filepath.Base(dir)}, nil
		}
	}
	dirs, _ := filepath.Glob(filepath.Join(root, "Versions", "*", PlayerExe))
	if len(dirs) == 0 {
		return Install{Root: root}, errors.New("roblox player not installed")
	}
	sort.Slice(dirs, func(i, j int) bool {
		a, _ := os.Stat(dirs[i])
		b, _ := os.Stat(dirs[j])
		return a.ModTime().After(b.ModTime())
	})
	dir := filepath.Dir(dirs[0])
	return Install{Root: root, VersionDir: dir, Player: dirs[0], Version: filepath.Base(dir)}, nil
}

func playerFromProtocol() string {
	for _, key := range []string{`Software\Classes\roblox-player\shell\open\command`, `Software\Classes\roblox\shell\open\command`} {
		k, err := registry.OpenKey(registry.CURRENT_USER, key, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		v, _, err := k.GetStringValue("")
		k.Close()
		if err != nil {
			continue
		}
		if m := quotedExe.FindStringSubmatch(v); m != nil {
			return m[1]
		}
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
