package app

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"toque/internal/i18n"
	"toque/internal/roblox"
	"toque/internal/winx"
)

var robloxSoundsDir = func() string {
	return filepath.Join(winx.UserLocalAppData(), "Temp", "Roblox", "sounds")
}

func dirBytes(dir string) int64 {
	var total int64
	filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if info, err := d.Info(); err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}

func emptyDir(dir string) (int64, error) {
	before := dirBytes(dir)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var errs []error
	for _, e := range entries {
		errs = append(errs, os.RemoveAll(filepath.Join(dir, e.Name())))
	}
	return before - dirBytes(dir), errors.Join(errs...)
}

func (s *Service) ClearRobloxCache() (int64, error) {
	if len(roblox.Running()) > 0 {
		return 0, errors.New(i18n.T("Cierra Roblox para borrar su caché de sonidos"))
	}
	freed, err := emptyDir(robloxSoundsDir())
	if d := loadLast[Diagnosis]("diagnosis.json"); d != nil {
		d.RobloxCacheMB = float64(dirBytes(robloxSoundsDir())) / (1 << 20)
		saveLast("diagnosis.json", *d)
	}
	return freed, err
}
