package roblox

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"

	"toque/internal/winx"
)

var reFlagName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,127}$`)

func ValidFlagName(name string) bool {
	return reFlagName.MatchString(name)
}

func FlagsPath(versionDir string) string {
	return filepath.Join(versionDir, "ClientSettings", "ClientAppSettings.json")
}

func ReadFlags(versionDir string) (map[string]any, error) {
	b, err := os.ReadFile(FlagsPath(versionDir))
	if errors.Is(err, os.ErrNotExist) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	if len(b) == 0 {
		return out, nil
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func WriteFlags(versionDir string, flags map[string]any) error {
	b, err := json.MarshalIndent(flags, "", "  ")
	if err != nil {
		return err
	}
	path := FlagsPath(versionDir)
	return winx.AsUser(func() error {
		if len(flags) == 0 {
			err := os.Remove(path)
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		return os.WriteFile(path, b, 0o644)
	})
}
