package tweaks

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var nicKeywords = map[string]bool{
	"*InterruptModeration": true,
	"*FlowControl":         true,
	"*EEE":                 true,
}

const gameAdapter = `$a = (Get-NetRoute -DestinationPrefix '0.0.0.0/0' | Sort-Object RouteMetric | Select-Object -First 1 | Get-NetAdapter).Name; `

type nicProperty struct {
	keyword string
	value   string
}

func (n nicProperty) ID() string { return "nic:" + n.keyword + "=" + n.value }
func (nicProperty) Kind() Kind   { return Relaunch }

func psOutput(script string) (string, error) {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func (n nicProperty) current() (string, error) {
	return psOutput(gameAdapter + fmt.Sprintf(`(Get-NetAdapterAdvancedProperty -Name $a -RegistryKeyword '%s').RegistryValue -join ','`, n.keyword))
}

func (n nicProperty) set(value string) error {
	if _, err := psOutput(gameAdapter + fmt.Sprintf(`Set-NetAdapterAdvancedProperty -Name $a -RegistryKeyword '%s' -RegistryValue '%s'`, n.keyword, value)); err != nil {
		return err
	}
	return waitOnline(40 * time.Second)
}

func waitOnline(timeout time.Duration) error {
	time.Sleep(2 * time.Second)
	for deadline := time.Now().Add(timeout); time.Now().Before(deadline); time.Sleep(500 * time.Millisecond) {
		if c, err := net.DialTimeout("tcp", "1.1.1.1:443", time.Second); err == nil {
			c.Close()
			return nil
		}
	}
	return errors.New("la red no volvió después de reiniciar la tarjeta")
}

func (n nicProperty) State(Context) State {
	cur, err := n.current()
	if err != nil || cur == "" {
		return State{Detail: "propiedad no disponible en esta tarjeta"}
	}
	return State{Applicable: true, Applied: cur == n.value, Detail: n.keyword + "=" + cur}
}

func (n nicProperty) Snapshot(Context) (Undo, error) {
	cur, err := n.current()
	if err != nil {
		return nil, err
	}
	if cur == "" {
		return nil, errors.New("propiedad no disponible en esta tarjeta")
	}
	return Undo{"prev": cur}, nil
}

func (n nicProperty) Apply(Context) error {
	return n.set(n.value)
}

func (n nicProperty) Revert(_ Context, u Undo) error {
	return n.set(u["prev"])
}

func digitsOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
