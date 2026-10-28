package tweaks

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"strings"

	"toque/internal/i18n"
	"toque/internal/roblox"
)

const (
	maxFlagValue   = 4096
	maxFlagsChange = 2000
)

func FlagID(name, value string) string {
	return "flag:" + name + "=" + url.QueryEscape(value)
}

func parseFlagID(id string) (map[string]any, error) {
	flags := map[string]any{}
	for _, part := range strings.Split(strings.TrimPrefix(id, "flag:"), ",") {
		k, raw, ok := strings.Cut(part, "=")
		v, err := url.QueryUnescape(raw)
		if !ok || err != nil || !roblox.ValidFlagName(k) {
			return nil, errors.New("bad flag tweak: " + part)
		}
		flags[k] = flagValue(v)
	}
	return flags, nil
}

func flagValue(v string) string {
	switch strings.ToLower(v) {
	case "true":
		return "True"
	case "false":
		return "False"
	}
	return v
}

func snapshotFlag(cur map[string]any, k string) string {
	old, ok := cur[k]
	if !ok {
		return ""
	}
	b, _ := json.Marshal(old)
	return string(b)
}

func restoreFlag(cur map[string]any, k, old string) {
	if old == "" {
		delete(cur, k)
		return
	}
	var v any
	json.Unmarshal([]byte(old), &v)
	cur[k] = v
}

func FlagsManaged(managed map[string]Undo, id string) bool {
	if _, ok := managed[id]; ok {
		return true
	}
	flags, err := parseFlagID(id)
	if err != nil {
		return false
	}
	for k, v := range flags {
		if _, ok := managed[FlagID(k, fmt.Sprint(v))]; !ok {
			return false
		}
	}
	return true
}

func UpdateFlags(ctx Context, set map[string]string, clear []string) error {
	if ctx.Install.VersionDir == "" {
		return errors.New(i18n.T("Roblox no está instalado"))
	}
	if len(set)+len(clear) > maxFlagsChange {
		return fmt.Errorf(i18n.T("como mucho %d flags de una vez"), maxFlagsChange)
	}
	touched := map[string]bool{}
	for k, v := range set {
		if !roblox.ValidFlagName(k) {
			return fmt.Errorf(i18n.T("%q no es un nombre de flag válido"), k)
		}
		if len(v) > maxFlagValue {
			return fmt.Errorf(i18n.T("el valor de %s es demasiado largo"), k)
		}
		touched[k] = true
	}
	for _, k := range clear {
		if !roblox.ValidFlagName(k) {
			return fmt.Errorf(i18n.T("%q no es un nombre de flag válido"), k)
		}
		touched[k] = true
	}
	journalMu.Lock()
	defer journalMu.Unlock()
	j, err := loadJournal()
	if err != nil {
		return err
	}
	before := maps.Clone(j.Applied)
	cur, err := roblox.ReadFlags(ctx.Install.VersionDir)
	if err != nil {
		return err
	}
	releaseFlags(j, cur, touched)
	for _, k := range clear {
		delete(cur, k)
	}
	for k, v := range set {
		v = flagValue(v)
		j.Applied[FlagID(k, v)] = Undo{k: snapshotFlag(cur, k)}
		cur[k] = v
	}
	if err := j.save(); err != nil {
		return err
	}
	if err := roblox.WriteFlags(ctx.Install.VersionDir, cur); err != nil {
		j.Applied = before
		j.save()
		return err
	}
	return nil
}

func releaseFlags(j journal, cur map[string]any, touched map[string]bool) {
	for _, id := range j.ids() {
		if !strings.HasPrefix(id, "flag:") {
			continue
		}
		flags, err := parseFlagID(id)
		if err != nil || !touchesAny(flags, touched) {
			continue
		}
		undo := j.Applied[id]
		for k := range flags {
			restoreFlag(cur, k, undo[k])
		}
		delete(j.Applied, id)
		for k, v := range flags {
			if touched[k] {
				continue
			}
			value := fmt.Sprint(v)
			j.Applied[FlagID(k, value)] = Undo{k: snapshotFlag(cur, k)}
			cur[k] = value
		}
	}
}

func touchesAny(flags map[string]any, touched map[string]bool) bool {
	for k := range flags {
		if touched[k] {
			return true
		}
	}
	return false
}
