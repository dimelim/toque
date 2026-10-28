package tweaks

import (
	"errors"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"

	"toque/internal/winx"
)

const (
	directXKey   = `Software\Microsoft\DirectX\UserGpuPreferences`
	directXValue = "DirectXUserGlobalSettings"
	swapUpgrade  = "SwapEffectUpgradeEnable"
	windowedID   = "windowed-games"
)

type windowedGames struct{}

func (windowedGames) ID() string { return windowedID }
func (windowedGames) Kind() Kind { return Relaunch }

func windows11() bool {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	build, _, err := k.GetStringValue("CurrentBuild")
	n, _ := strconv.Atoi(build)
	return err == nil && n >= 22621
}

func readDirectX() (string, bool, error) {
	root, err := winx.UserRoot()
	if err != nil {
		return "", false, err
	}
	k, err := registry.OpenKey(root, directXKey, registry.QUERY_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	defer k.Close()
	v, _, err := k.GetStringValue(directXValue)
	if errors.Is(err, registry.ErrNotExist) {
		return "", false, nil
	}
	return v, err == nil, err
}

func writeDirectX(v string) error {
	root, err := winx.UserRoot()
	if err != nil {
		return err
	}
	k, _, err := registry.CreateKey(root, directXKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringValue(directXValue, v)
}

func removeDirectX() error {
	root, err := winx.UserRoot()
	if err != nil {
		return err
	}
	k, err := registry.OpenKey(root, directXKey, registry.SET_VALUE)
	if err != nil {
		return ignoreMissing(err)
	}
	defer k.Close()
	return ignoreMissing(k.DeleteValue(directXValue))
}

func directXSetting(settings, name string) string {
	for _, part := range strings.Split(settings, ";") {
		if k, v, ok := strings.Cut(part, "="); ok && k == name {
			return v
		}
	}
	return ""
}

func withDirectXSetting(settings, name, value string) string {
	var parts []string
	found := false
	for _, part := range strings.Split(settings, ";") {
		if part == "" {
			continue
		}
		if k, _, ok := strings.Cut(part, "="); ok && k == name {
			if value == "" {
				continue
			}
			part, found = name+"="+value, true
		}
		parts = append(parts, part)
	}
	if !found && value != "" {
		parts = append(parts, name+"="+value)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ";") + ";"
}

func (windowedGames) State(Context) State {
	cur, _, err := readDirectX()
	return State{Applicable: windows11() && err == nil, Applied: directXSetting(cur, swapUpgrade) == "1"}
}

func (windowedGames) Snapshot(Context) (Undo, error) {
	cur, ok, err := readDirectX()
	if err != nil {
		return nil, err
	}
	u := Undo{"setting": directXSetting(cur, swapUpgrade)}
	if !ok {
		u["value"] = absent
	}
	return u, nil
}

func (windowedGames) Apply(Context) error {
	cur, _, err := readDirectX()
	if err != nil {
		return err
	}
	return writeDirectX(withDirectXSetting(cur, swapUpgrade, "1"))
}

func (windowedGames) Revert(_ Context, u Undo) error {
	prev, ok := u["setting"]
	if !ok {
		return nil
	}
	cur, _, err := readDirectX()
	if err != nil {
		return err
	}
	next := withDirectXSetting(cur, swapUpgrade, prev)
	if next == "" && u["value"] == absent {
		return removeDirectX()
	}
	return writeDirectX(next)
}

func WindowedGames() Tweak { return windowedGames{} }

func init() {
	Register(windowedGames{})
}
