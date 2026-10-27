package roblox

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"toque/internal/winx"
)

var AllowedFlags = map[string]bool{
	"DFIntCSGLevelOfDetailSwitchingDistance":    true,
	"DFIntCSGLevelOfDetailSwitchingDistanceL12": true,
	"DFIntCSGLevelOfDetailSwitchingDistanceL23": true,
	"DFIntCSGLevelOfDetailSwitchingDistanceL34": true,
	"FFlagHandleAltEnterFullscreenManually":     true,
	"DFFlagTextureQualityOverrideEnabled":       true,
	"DFIntTextureQualityOverride":               true,
	"FIntDebugForceMSAASamples":                 true,
	"DFFlagDisableDPIScale":                     true,
	"FFlagDebugGraphicsPreferD3D11":             true,
	"FFlagDebugSkyGray":                         true,
	"DFFlagDebugPauseVoxelizer":                 true,
	"DFIntDebugFRMQualityLevelOverride":         true,
	"FIntFRMMaxGrassDistance":                   true,
	"FIntFRMMinGrassDistance":                   true,
	"FFlagDebugGraphicsPreferVulkan":            true,
	"FFlagDebugGraphicsPreferOpenGL":            true,
	"FIntGrassMovementReducedMotionFactor":      true,
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
	for k := range flags {
		if !AllowedFlags[k] {
			return errors.New("flag not in Roblox allowlist: " + k)
		}
	}
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
