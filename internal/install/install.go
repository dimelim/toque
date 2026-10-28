package install

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"toque/internal/i18n"
	"toque/internal/winx"
)

//go:embed payload
var payload embed.FS

const uninstallKey = `Software\Microsoft\Windows\CurrentVersion\Uninstall\Toque`

type State struct {
	Installed        bool   `json:"installed"`
	InstalledVersion string `json:"installedVersion"`
	Dir              string `json:"dir"`
	RunningInstalled bool   `json:"runningInstalled"`
	Version          string `json:"version"`
	Newer            bool   `json:"newer"`
}

func Dir() string {
	pf, err := windows.KnownFolderPath(windows.FOLDERID_ProgramFiles, 0)
	if err != nil {
		pf = os.Getenv("ProgramFiles")
	}
	return filepath.Join(pf, "Toque")
}

func startMenuShortcut() string {
	p, err := windows.KnownFolderPath(windows.FOLDERID_CommonPrograms, 0)
	if err != nil {
		p = filepath.Join(os.Getenv("ProgramData"), `Microsoft\Windows\Start Menu\Programs`)
	}
	return filepath.Join(p, "Toque.lnk")
}

func HasDesktopShortcut() bool {
	d := desktopShortcut()
	if d == "" {
		return false
	}
	_, err := os.Stat(d)
	return err == nil
}

func desktopShortcut() string {
	p, err := windows.KnownFolderPath(windows.FOLDERID_PublicDesktop, 0)
	if err != nil {
		return ""
	}
	return filepath.Join(p, "Toque.lnk")
}

func Current(version string) State {
	st := State{Dir: Dir(), Version: version}
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, uninstallKey, registry.QUERY_VALUE)
	if err == nil {
		st.InstalledVersion, _, _ = k.GetStringValue("DisplayVersion")
		k.Close()
	}
	if _, err := os.Stat(filepath.Join(st.Dir, "Toque.exe")); err == nil && st.InstalledVersion != "" {
		st.Installed = true
	}
	exe, _ := os.Executable()
	st.RunningInstalled = strings.EqualFold(filepath.Clean(filepath.Dir(exe)), filepath.Clean(st.Dir))
	st.Newer = st.Installed && Compare(version, st.InstalledVersion) > 0
	return st
}

func Compare(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		var x, y int
		if i < len(pa) {
			fmt.Sscan(pa[i], &x)
		}
		if i < len(pb) {
			fmt.Sscan(pb[i], &y)
		}
		if x != y {
			if x > y {
				return 1
			}
			return -1
		}
	}
	return 0
}

func stopRunning(except uint32) {
	for _, p := range winx.FindByName("Toque.exe") {
		if p.PID != except {
			winx.Terminate(p.PID)
		}
	}
	time.Sleep(600 * time.Millisecond)
}

func sha(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha256.New()
	io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil))
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	tmp := dst + ".new"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	out.Close()
	os.Remove(dst + ".old")
	if _, err := os.Stat(dst); err == nil {
		if err := os.Rename(dst, dst+".old"); err != nil {
			return err
		}
	}
	return os.Rename(tmp, dst)
}

func writePayload(dir string) error {
	return fs.WalkDir(payload, "payload", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := payload.ReadFile(path)
		if err != nil {
			return err
		}
		dst := filepath.Join(dir, filepath.Base(path))
		tmp := dst + ".new"
		if err := os.WriteFile(tmp, b, 0o755); err != nil {
			return err
		}
		os.Remove(dst + ".old")
		if _, err := os.Stat(dst); err == nil {
			os.Rename(dst, dst+".old")
		}
		return os.Rename(tmp, dst)
	})
}

func powershell(script string) error {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	return cmd.Run()
}

func shortcut(lnk, target string) error {
	if lnk == "" {
		return nil
	}
	esc := func(s string) string { return strings.ReplaceAll(s, "'", "''") }
	return powershell(fmt.Sprintf("$s=(New-Object -ComObject WScript.Shell).CreateShortcut('%s');$s.TargetPath='%s';$s.WorkingDirectory='%s';$s.IconLocation='%s,0';$s.Description='Toque';$s.Save()",
		esc(lnk), esc(target), esc(filepath.Dir(target)), esc(target)))
}

func Install(version string, desktop bool) (string, error) {
	if !winx.IsElevated() {
		return "", errors.New(i18n.T("se necesitan permisos de administrador"))
	}
	src, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := Dir()
	stopRunning(uint32(os.Getpid()))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	target := filepath.Join(dir, "Toque.exe")
	if !strings.EqualFold(filepath.Clean(src), filepath.Clean(target)) {
		if err := copyFile(src, target); err != nil {
			return "", fmt.Errorf(i18n.T("copiar Toque.exe (%w)"), err)
		}
	}
	if err := writePayload(dir); err != nil {
		return "", fmt.Errorf(i18n.T("copiar componentes (%w)"), err)
	}
	if sha(target) == "" {
		return "", errors.New(i18n.T("la copia de Toque.exe está vacía"))
	}
	shortcut(startMenuShortcut(), target)
	if desktop {
		shortcut(desktopShortcut(), target)
	}
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, uninstallKey, registry.SET_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()
	k.SetStringValue("DisplayName", "Toque")
	k.SetStringValue("DisplayVersion", version)
	k.SetStringValue("Publisher", "Toque")
	k.SetStringValue("DisplayIcon", target+",0")
	k.SetStringValue("InstallLocation", dir)
	k.SetStringValue("UninstallString", `"`+target+`" --uninstall`)
	k.SetStringValue("QuietUninstallString", `"`+target+`" --uninstall --quiet`)
	k.SetDWordValue("NoModify", 1)
	k.SetDWordValue("NoRepair", 1)
	k.SetDWordValue("EstimatedSize", uint32(dirSizeKB(dir)))
	return target, nil
}

func dirSizeKB(dir string) int64 {
	var total int64
	filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if info, err := d.Info(); err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total / 1024
}

func Uninstall() error {
	if !winx.IsElevated() {
		return errors.New(i18n.T("se necesitan permisos de administrador"))
	}
	stopRunning(uint32(os.Getpid()))
	os.Remove(startMenuShortcut())
	if d := desktopShortcut(); d != "" {
		os.Remove(d)
	}
	registry.DeleteKey(registry.LOCAL_MACHINE, uninstallKey)
	data := filepath.Join(winx.ProgramData(), "Toque")
	webview := filepath.Join(winx.LocalAppData(), "Toque")
	os.RemoveAll(data)
	os.RemoveAll(webview)
	dir := Dir()
	exe, _ := os.Executable()
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		p := filepath.Join(dir, e.Name())
		if strings.EqualFold(p, exe) {
			continue
		}
		os.RemoveAll(p)
	}
	deleteOnReboot(exe)
	deleteOnReboot(dir)
	cmdExe := filepath.Join(os.Getenv("SystemRoot"), "System32", "cmd.exe")
	line := fmt.Sprintf(`"%s" /c ping 127.0.0.1 -n 3 > nul & rmdir /s /q "%s" & rmdir /s /q "%s" & rmdir /s /q "%s"`, cmdExe, dir, data, webview)
	const noWindow, detached, breakaway = 0x08000000, 0x00000008, 0x01000000
	for _, flags := range []uint32{noWindow | detached | breakaway, noWindow | detached} {
		cmd := exec.Command(cmdExe)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: flags, CmdLine: line}
		if err := cmd.Start(); err == nil {
			return nil
		}
	}
	return nil
}

var procMoveFileExW = windows.NewLazySystemDLL("kernel32.dll").NewProc("MoveFileExW")

func deleteOnReboot(path string) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return
	}
	const delayUntilReboot = 0x4
	procMoveFileExW.Call(uintptr(unsafe.Pointer(p)), 0, delayUntilReboot)
}

func CleanupOld() {
	dir := Dir()
	matches, _ := filepath.Glob(filepath.Join(dir, "*.old"))
	for _, m := range matches {
		os.Remove(m)
	}
}
