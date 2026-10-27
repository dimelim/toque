package app

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strconv"

	"toque/internal/roblox"
	"toque/internal/tweaks"
)

type FlagEntry struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Allowed bool   `json:"allowed"`
}

type FlagsInfo struct {
	Flags     []FlagEntry `json:"flags"`
	Allowlist []string    `json:"allowlist"`
}

func (s *Service) Flags() (FlagsInfo, error) {
	info := FlagsInfo{Flags: []FlagEntry{}, Allowlist: slices.Sorted(maps.Keys(roblox.AllowedFlags))}
	dir := s.context().Install.VersionDir
	if dir == "" {
		return info, errors.New("Roblox no está instalado")
	}
	cur, err := roblox.ReadFlags(dir)
	if err != nil {
		return info, fmt.Errorf("no pude leer las flags de Roblox: %w", err)
	}
	for k, v := range cur {
		info.Flags = append(info.Flags, FlagEntry{Name: k, Value: flagText(v), Allowed: roblox.AllowedFlags[k]})
	}
	sort.Slice(info.Flags, func(i, j int) bool { return info.Flags[i].Name < info.Flags[j].Name })
	return info, nil
}

func (s *Service) UpdateFlags(set map[string]string, clear []string) (FlagsInfo, error) {
	if err := s.idle(); err != nil {
		return FlagsInfo{}, err
	}
	if err := tweaks.UpdateFlags(s.context(), set, clear); err != nil {
		return FlagsInfo{}, err
	}
	return s.Flags()
}

func flagText(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case bool:
		if x {
			return "True"
		}
		return "False"
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	}
	return fmt.Sprint(v)
}
