package bench

import (
	"encoding/json"
	"os"
	"path/filepath"

	"toque/internal/roblox"
	"toque/internal/tweaks"
	"toque/internal/winx"
)

type marker struct {
	Tweak      string `json:"tweak"`
	WasApplied bool   `json:"wasApplied"`
}

func markerPath() string {
	return filepath.Join(winx.DataDir(), "experiment-active.json")
}

func writeMarker(id string, wasApplied bool) {
	b, _ := json.Marshal(marker{Tweak: id, WasApplied: wasApplied})
	os.WriteFile(markerPath(), b, 0o644)
}

func clearMarker() {
	os.Remove(markerPath())
}

func RecoverInterrupted(inst roblox.Install) (string, error) {
	b, err := os.ReadFile(markerPath())
	if err != nil {
		return "", nil
	}
	var m marker
	if json.Unmarshal(b, &m) != nil {
		clearMarker()
		return "", nil
	}
	t, err := tweaks.Get(m.Tweak)
	if err != nil {
		clearMarker()
		return "", err
	}
	if t.Kind() != tweaks.Live && len(roblox.Running()) > 0 {
		return "", nil
	}
	ctx := tweaks.Context{Install: inst}
	if m.WasApplied {
		err = tweaks.Apply(ctx, t)
	} else {
		err = tweaks.Revert(ctx, t)
	}
	if err == nil {
		clearMarker()
	}
	return m.Tweak, err
}
