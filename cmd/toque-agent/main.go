package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"

	"toque/internal/catalog"
	"toque/internal/session"
	"toque/internal/tray"
)

func shellOpen(path string) {
	verb, _ := windows.UTF16PtrFromString("open")
	file, _ := windows.UTF16PtrFromString(path)
	windows.ShellExecute(0, verb, file, nil, nil, windows.SW_SHOWNORMAL)
}

func main() {
	ok, release := session.Exclusive(`Local\ToqueSession`)
	if !ok {
		return
	}
	defer release()
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	ui := filepath.Join(dir, "Toque.exe")
	icon := exe
	if _, err := os.Stat(ui); err == nil {
		icon = ui
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctl := session.NewController(catalog.LiveProven())
	t := &tray.Tray{}
	ctl.OnChange = func(session.Status) { t.Refresh() }
	t.Tooltip = func() string {
		if st := ctl.Status(); st.Active {
			return "Toque · optimizando Roblox"
		}
		return "Toque · esperando a Roblox"
	}
	t.Status = func() string {
		st := ctl.Status()
		if !st.Active {
			return "Esperando a que abras Roblox"
		}
		if len(st.Applied) == 0 {
			return "Roblox abierto"
		}
		return "Activo con " + strings.Join(catalog.Titles(st.Applied), ", ")
	}
	t.OnOpen = func() { shellOpen(ui) }
	t.OnQuit = cancel
	done := make(chan struct{})
	go func() {
		ctl.Run(ctx)
		close(done)
	}()
	t.Run(icon)
	cancel()
	<-done
}
