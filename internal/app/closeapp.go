package app

import (
	"errors"
	"fmt"
	"strings"

	"toque/internal/eco"
	"toque/internal/winx"
)

var neverClose = map[string]bool{
	"explorer.exe": true, "msmpeng.exe": true, "securityhealthsystray.exe": true, "searchhost.exe": true,
	"shellexperiencehost.exe": true, "startmenuexperiencehost.exe": true, "textinputhost.exe": true,
	"lockapp.exe": true, "applicationframehost.exe": true, "sihost.exe": true, "ctfmon.exe": true,
	"toque.exe": true, "exitlag.exe": true, "msedgewebview2.exe": true,
}

func appExes(name string) map[string]bool {
	exes := map[string]bool{}
	for _, exe := range eco.Exes(name) {
		exes[exe] = true
	}
	if len(exes) == 0 {
		exes[strings.ToLower(name)+".exe"] = true
	}
	return exes
}

func (s *Service) CloseApp(name string) (int, error) {
	exes := appExes(name)
	for exe := range exes {
		if neverClose[exe] || quietProcs[exe] {
			return 0, errors.New("Toque no cierra ese programa")
		}
	}
	session, _ := winx.SessionOf(ownPID())
	procs, err := winx.Processes()
	if err != nil {
		return 0, err
	}
	pids := map[uint32]bool{}
	for _, p := range procs {
		if !exes[strings.ToLower(p.Name)] || p.PID == ownPID() {
			continue
		}
		if sid, ok := winx.SessionOf(p.PID); ok && sid == session {
			pids[p.PID] = true
		}
	}
	if len(pids) == 0 {
		return 0, fmt.Errorf("%s ya no está abierto", name)
	}
	n := winx.CloseWindowsOf(pids)
	if n == 0 {
		return 0, fmt.Errorf("%s no tiene ventanas abiertas. Ciérralo desde su icono junto al reloj", name)
	}
	return n, nil
}
