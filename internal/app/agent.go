package app

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"

	"toque/internal/winx"
)

const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`

func stopAgent() {
	if winx.CloseWindowByClass("ToqueAgentTray") {
		for i := 0; i < 20 && len(winx.FindByName("ToqueAgent.exe")) > 0; i++ {
			time.Sleep(100 * time.Millisecond)
		}
	}
	for _, p := range winx.FindByName("ToqueAgent.exe") {
		winx.Terminate(p.PID)
	}
}

func startAgent() error {
	path := agentPath()
	if _, err := os.Stat(path); err != nil {
		return err
	}
	if AgentRunning() {
		return nil
	}
	_, err := winx.LaunchAsUser(path, "", filepath.Dir(path))
	return err
}

func AgentRunning() bool {
	return len(winx.FindByName("ToqueAgent.exe")) > 0
}

func agentPath() string {
	exe, _ := os.Executable()
	return filepath.Join(filepath.Dir(exe), "ToqueAgent.exe")
}

func (s *Service) AgentEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	v, _, err := k.GetStringValue("Toque")
	return err == nil && strings.Contains(strings.ToLower(v), "toqueagent.exe")
}

func (s *Service) SetAgentEnabled(on bool) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	path := agentPath()
	if !on {
		k.DeleteValue("Toque")
		stopAgent()
		s.StartSession()
		return nil
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	if err := k.SetStringValue("Toque", `"`+path+`"`); err != nil {
		return err
	}
	s.StopSession()
	return startAgent()
}
