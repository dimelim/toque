package winx

import (
	"fmt"
	"runtime"

	"golang.org/x/sys/windows"
)

func AsUser(fn func() error) error {
	if !IsElevated() {
		return fn()
	}
	primary, err := shellToken()
	if err != nil {
		return fmt.Errorf("no encuentro la sesión del usuario: %w", err)
	}
	defer primary.Close()
	var imp windows.Token
	if err := windows.DuplicateTokenEx(primary, windows.TOKEN_QUERY|windows.TOKEN_IMPERSONATE, nil, windows.SecurityImpersonation, windows.TokenImpersonation, &imp); err != nil {
		return err
	}
	defer imp.Close()
	runtime.LockOSThread()
	if err := windows.SetThreadToken(nil, imp); err != nil {
		runtime.UnlockOSThread()
		return err
	}
	err = fn()
	if windows.RevertToSelf() == nil {
		runtime.UnlockOSThread()
	}
	return err
}
