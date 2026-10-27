package app

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/sys/windows"

	"toque/internal/roblox"
	"toque/internal/winx"
)

const (
	errorLogMax      = 256 * 1024
	reportRobloxLogs = 3
)

func errorLogPath() string { return filepath.Join(winx.DataDir(), "errores.log") }

func (s *Service) NoteError(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	path := errorLogPath()
	if info, err := os.Stat(path); err == nil && info.Size() > errorLogMax {
		os.Rename(path, path+".old")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s  v%s  %s\n", time.Now().Format(time.RFC3339), Version, strings.ReplaceAll(text, "\n", " "))
}

func (s *Service) ExportReport() (string, error) {
	desktop, err := windows.KnownFolderPath(windows.FOLDERID_Desktop, 0)
	if err != nil {
		return "", fmt.Errorf("no encuentro tu escritorio: %w", err)
	}
	path := filepath.Join(desktop, "Toque-informe-"+time.Now().Format("20060102-150405")+".zip")
	if err := writeReport(path); err != nil {
		os.Remove(path)
		return "", fmt.Errorf("no pude crear el informe: %w", err)
	}
	winx.LaunchAsUser(filepath.Join(os.Getenv("WINDIR"), "explorer.exe"), `/select,"`+path+`"`, "")
	return path, nil
}

func writeReport(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	z := zip.NewWriter(f)
	inst, findErr := roblox.Find()
	info := map[string]any{
		"toque":    Version,
		"creado":   time.Now().Format(time.RFC3339),
		"admin":    winx.IsElevated(),
		"sistema":  ReadSystem(),
		"roblox":   inst,
		"robloxOk": findErr == nil,
	}
	if b, err := json.MarshalIndent(info, "", "  "); err == nil {
		addBytes(z, "info.json", b)
	}
	for _, name := range []string{"errores.log", "errores.log.old", "hud.json"} {
		addFile(z, name, filepath.Join(winx.DataDir(), name))
	}
	if inst.VersionDir != "" {
		addFile(z, "ClientAppSettings.json", filepath.Join(inst.VersionDir, "ClientSettings", "ClientAppSettings.json"))
	}
	for _, log := range newestPlayerLogs(inst.LogDir(), reportRobloxLogs) {
		addFile(z, "roblox/"+filepath.Base(log), log)
	}
	if err := z.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func newestPlayerLogs(dir string, n int) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	type logFile struct {
		path string
		mod  time.Time
	}
	var logs []logFile
	for _, e := range entries {
		if e.IsDir() || !roblox.IsPlayerLog(e.Name()) {
			continue
		}
		if info, err := e.Info(); err == nil {
			logs = append(logs, logFile{filepath.Join(dir, e.Name()), info.ModTime()})
		}
	}
	sort.Slice(logs, func(i, j int) bool { return logs[i].mod.After(logs[j].mod) })
	var out []string
	for i := 0; i < len(logs) && i < n; i++ {
		out = append(out, logs[i].path)
	}
	return out
}

func addBytes(z *zip.Writer, name string, b []byte) {
	if w, err := z.Create(name); err == nil {
		w.Write(b)
	}
}

func addFile(z *zip.Writer, name, src string) {
	f, err := os.Open(src)
	if err != nil {
		return
	}
	defer f.Close()
	if w, err := z.Create(name); err == nil {
		io.Copy(w, f)
	}
}
