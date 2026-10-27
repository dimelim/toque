package app

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"toque/internal/roblox"
	"toque/internal/tweaks"
)

type FlagEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (s *Service) Flags() ([]FlagEntry, error) {
	dir := s.context().Install.VersionDir
	if dir == "" {
		return []FlagEntry{}, errors.New("Roblox no está instalado")
	}
	cur, err := roblox.ReadFlags(dir)
	if err != nil {
		return []FlagEntry{}, fmt.Errorf("no pude leer las flags de Roblox: %w", err)
	}
	out := make([]FlagEntry, 0, len(cur))
	for k, v := range cur {
		out = append(out, FlagEntry{Name: k, Value: flagText(v)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *Service) UpdateFlags(set map[string]string, clear []string) ([]FlagEntry, error) {
	if err := s.idle(); err != nil {
		return nil, err
	}
	if err := tweaks.UpdateFlags(s.context(), set, clear); err != nil {
		return nil, err
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
