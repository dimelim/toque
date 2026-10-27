package winx

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                      = windows.NewLazySystemDLL("user32.dll")
	advapi32                    = windows.NewLazySystemDLL("advapi32.dll")
	procGetShellWindow          = user32.NewProc("GetShellWindow")
	procCreateProcessWithTokenW = advapi32.NewProc("CreateProcessWithTokenW")
)

func shellToken() (windows.Token, error) {
	hwnd, _, _ := procGetShellWindow.Call()
	if hwnd == 0 {
		return 0, errors.New("no shell window in this session")
	}
	var pid uint32
	if _, err := windows.GetWindowThreadProcessId(windows.HWND(hwnd), &pid); err != nil {
		return 0, err
	}
	proc, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(proc)
	var tok windows.Token
	if err := windows.OpenProcessToken(proc, windows.TOKEN_DUPLICATE|windows.TOKEN_QUERY, &tok); err != nil {
		return 0, err
	}
	defer tok.Close()
	var primary windows.Token
	access := uint32(windows.TOKEN_QUERY | windows.TOKEN_DUPLICATE | windows.TOKEN_ASSIGN_PRIMARY | windows.TOKEN_ADJUST_DEFAULT | windows.TOKEN_ADJUST_SESSIONID)
	if err := windows.DuplicateTokenEx(tok, access, nil, windows.SecurityImpersonation, windows.TokenPrimary, &primary); err != nil {
		return 0, err
	}
	return primary, nil
}

func ownToken() (windows.Token, error) {
	var tok windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY|windows.TOKEN_DUPLICATE|windows.TOKEN_IMPERSONATE, &tok)
	return tok, err
}

func LaunchAsUser(exe string, args string, dir string) (uint32, error) {
	cmdline := fmt.Sprintf("\"%s\" %s", exe, args)
	cmd, err := windows.UTF16PtrFromString(cmdline)
	if err != nil {
		return 0, err
	}
	var wdir *uint16
	if dir != "" {
		if wdir, err = windows.UTF16PtrFromString(dir); err != nil {
			return 0, err
		}
	}
	si := windows.StartupInfo{Cb: uint32(unsafe.Sizeof(windows.StartupInfo{}))}
	var pi windows.ProcessInformation
	flags := uint32(windows.CREATE_UNICODE_ENVIRONMENT | windows.CREATE_NEW_PROCESS_GROUP)

	if IsElevated() {
		tok, err := shellToken()
		if err != nil {
			return 0, fmt.Errorf("unelevated launch: %w", err)
		}
		defer tok.Close()
		var env *uint16
		if err := windows.CreateEnvironmentBlock(&env, tok, false); err != nil {
			return 0, err
		}
		defer windows.DestroyEnvironmentBlock(env)
		r, _, callErr := procCreateProcessWithTokenW.Call(
			uintptr(tok), 0, 0,
			uintptr(unsafe.Pointer(cmd)),
			uintptr(flags),
			uintptr(unsafe.Pointer(env)),
			uintptr(unsafe.Pointer(wdir)),
			uintptr(unsafe.Pointer(&si)),
			uintptr(unsafe.Pointer(&pi)),
		)
		if r == 0 {
			return 0, callErr
		}
	} else {
		tok, err := ownToken()
		if err != nil {
			return 0, err
		}
		defer tok.Close()
		var env *uint16
		if err := windows.CreateEnvironmentBlock(&env, tok, false); err != nil {
			return 0, err
		}
		defer windows.DestroyEnvironmentBlock(env)
		if err := windows.CreateProcess(nil, cmd, nil, nil, false, flags, env, wdir, &si, &pi); err != nil {
			return 0, err
		}
	}
	windows.CloseHandle(pi.Thread)
	windows.CloseHandle(pi.Process)
	return pi.ProcessId, nil
}
