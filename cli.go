package main

import (
	"os"
	"time"

	"golang.org/x/sys/windows"

	"toque/internal/install"
	"toque/internal/roblox"
	"toque/internal/tweaks"
)

const (
	mbYesNo       = 0x00000004
	mbIconQuestion = 0x00000020
	mbIconInfo    = 0x00000040
	mbIconWarning = 0x00000030
	idYes         = 6
)

func message(text string, flags uint32) int32 {
	t, _ := windows.UTF16PtrFromString(text)
	c, _ := windows.UTF16PtrFromString("Toque")
	r, _ := windows.MessageBox(0, t, c, flags)
	return r
}

func has(flag string) bool {
	for _, a := range os.Args[1:] {
		if a == flag {
			return true
		}
	}
	return false
}

func revertAll() error {
	roblox.CloseAll(5 * time.Second)
	ctx := tweaks.Context{}
	if inst, err := roblox.Find(); err == nil {
		ctx.Install = inst
	}
	return tweaks.OpenJournal().RevertAll(ctx)
}

func runCLI() bool {
	quiet := has("--quiet")
	switch {
	case has("--uninstall"):
		if !quiet && message("¿Desinstalar Toque?\n\nSe desharán todos los cambios que Toque hizo en Windows y en Roblox.", mbYesNo|mbIconQuestion) != idYes {
			return true
		}
		if err := revertAll(); err != nil && !quiet {
			message("Algunos cambios no se pudieron revertir:\n"+err.Error(), mbIconWarning)
		}
		if err := install.Uninstall(true); err != nil {
			if !quiet {
				message("No se pudo desinstalar: "+err.Error(), mbIconWarning)
			}
			os.Exit(1)
		}
		if !quiet {
			message("Toque se desinstaló y todo volvió a como estaba.", mbIconInfo)
		}
		return true
	case has("--revert-all"):
		if err := revertAll(); err != nil {
			if !quiet {
				message("Algunos cambios no se pudieron revertir:\n"+err.Error(), mbIconWarning)
			}
			os.Exit(1)
		}
		return true
	}
	install.CleanupOld()
	return false
}
