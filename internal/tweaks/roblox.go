package tweaks

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"toque/internal/roblox"
)

type flagSet struct {
	id    string
	flags map[string]any
}

func (f flagSet) ID() string { return f.id }
func (flagSet) Kind() Kind   { return Relaunch }

func (f flagSet) State(c Context) State {
	cur, err := roblox.ReadFlags(c.Install.VersionDir)
	if err != nil {
		return State{Detail: err.Error()}
	}
	for k, v := range f.flags {
		if fmt.Sprint(cur[k]) != fmt.Sprint(v) {
			return State{Applicable: true, Applied: false}
		}
	}
	return State{Applicable: true, Applied: true}
}

func (f flagSet) Apply(c Context) (Undo, error) {
	if len(roblox.Running()) > 0 {
		return nil, errors.New("close roblox first")
	}
	cur, err := roblox.ReadFlags(c.Install.VersionDir)
	if err != nil {
		return nil, err
	}
	u := Undo{}
	for k, v := range f.flags {
		if old, ok := cur[k]; ok {
			b, _ := json.Marshal(old)
			u[k] = string(b)
		} else {
			u[k] = ""
		}
		cur[k] = v
	}
	return u, roblox.WriteFlags(c.Install.VersionDir, cur)
}

func (f flagSet) Revert(c Context, u Undo) error {
	if len(roblox.Running()) > 0 {
		return errors.New("close roblox first")
	}
	cur, err := roblox.ReadFlags(c.Install.VersionDir)
	if err != nil {
		return err
	}
	for k, old := range u {
		if old == "" {
			delete(cur, k)
			continue
		}
		var v any
		json.Unmarshal([]byte(old), &v)
		cur[k] = v
	}
	return roblox.WriteFlags(c.Install.VersionDir, cur)
}

type basicSetting struct {
	kind  string
	name  string
	value string
}

func (b basicSetting) ID() string { return "setting:" + b.name + "=" + b.value }
func (basicSetting) Kind() Kind   { return Relaunch }

func (b basicSetting) read(c Context) (string, bool, *roblox.BasicSettings, error) {
	st, err := roblox.LoadBasicSettings(c.Install.SettingsPath())
	if err != nil {
		return "", false, nil, err
	}
	switch b.kind {
	case "int":
		v, ok := st.Int(b.name)
		return strconv.Itoa(v), ok, st, nil
	case "token":
		v, ok := st.Token(b.name)
		return strconv.Itoa(v), ok, st, nil
	default:
		v, ok := st.Bool(b.name)
		return strconv.FormatBool(v), ok, st, nil
	}
}

func (b basicSetting) write(st *roblox.BasicSettings, value string) error {
	ok := false
	switch b.kind {
	case "int":
		n, _ := strconv.Atoi(value)
		ok = st.SetInt(b.name, n)
	case "token":
		n, _ := strconv.Atoi(value)
		ok = st.SetToken(b.name, n)
	default:
		ok = st.SetBool(b.name, value == "true")
	}
	if !ok {
		return errors.New("setting not found: " + b.name)
	}
	return st.Save()
}

func (b basicSetting) State(c Context) State {
	cur, ok, _, err := b.read(c)
	if err != nil || !ok {
		return State{Detail: "setting unavailable"}
	}
	return State{Applicable: true, Applied: cur == b.value, Detail: b.name + "=" + cur}
}

func (b basicSetting) Apply(c Context) (Undo, error) {
	if len(roblox.Running()) > 0 {
		return nil, errors.New("close roblox first")
	}
	cur, ok, st, err := b.read(c)
	if err != nil || !ok {
		return nil, errors.New("setting unavailable: " + b.name)
	}
	return Undo{"prev": cur}, b.write(st, b.value)
}

func (b basicSetting) Revert(c Context, u Undo) error {
	if len(roblox.Running()) > 0 {
		return errors.New("close roblox first")
	}
	_, _, st, err := b.read(c)
	if err != nil {
		return err
	}
	return b.write(st, u["prev"])
}

var settingKinds = map[string]string{
	"FramerateCap":             "int",
	"SavedQualityLevel":        "token",
	"GraphicsOptimizationMode": "token",
	"GraphicsQualityLevel":     "int",
	"Fullscreen":               "bool",
	"PerformanceStatsVisible":  "bool",
}

func parametric(id string) (Tweak, error) {
	switch {
	case strings.HasPrefix(id, "power-plan:"):
		return powerPlan{target: strings.TrimPrefix(id, "power-plan:")}, nil
	case strings.HasPrefix(id, "setting:"):
		name, value, ok := strings.Cut(strings.TrimPrefix(id, "setting:"), "=")
		kind, known := settingKinds[name]
		if !ok || !known {
			return nil, errors.New("bad setting tweak")
		}
		return basicSetting{kind: kind, name: name, value: value}, nil
	case strings.HasPrefix(id, "flag:"):
		flags := map[string]any{}
		for _, part := range strings.Split(strings.TrimPrefix(id, "flag:"), ",") {
			k, v, ok := strings.Cut(part, "=")
			if !ok || !roblox.AllowedFlags[k] {
				return nil, errors.New("bad flag tweak: " + part)
			}
			flags[k] = parseFlagValue(v)
		}
		return flagSet{id: id, flags: flags}, nil
	}
	return nil, errors.New("not parametric")
}

func parseFlagValue(v string) any {
	switch strings.ToLower(v) {
	case "true":
		return "True"
	case "false":
		return "False"
	}
	return v
}
