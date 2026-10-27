package winx

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func LocalAppData() string {
	if p, err := windows.KnownFolderPath(windows.FOLDERID_LocalAppData, 0); err == nil {
		return p
	}
	return os.Getenv("LOCALAPPDATA")
}

func ProgramData() string {
	if p, err := windows.KnownFolderPath(windows.FOLDERID_ProgramData, 0); err == nil {
		return p
	}
	return os.Getenv("ProgramData")
}

func DataDir(parts ...string) string {
	p := filepath.Join(append([]string{ProgramData(), "Toque"}, parts...)...)
	os.MkdirAll(p, 0o755)
	return p
}
