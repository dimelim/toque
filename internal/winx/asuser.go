package winx

import (
	"fmt"
	"runtime"

	"golang.org/x/sys/windows"

	"toque/internal/i18n"
)

func shellImpersonation() (windows.Token, error) {
	primary, err := shellToken()
	if err != nil {
		return 0, fmt.Errorf(i18n.T("no encuentro la sesión del usuario (%w)"), err)
	}
	defer primary.Close()
	var imp windows.Token
	err = windows.DuplicateTokenEx(primary, windows.TOKEN_QUERY|windows.TOKEN_IMPERSONATE, nil, windows.SecurityImpersonation, windows.TokenImpersonation, &imp)
	return imp, err
}

func AsUser(fn func() error) error {
	if !IsElevated() {
		return fn()
	}
	imp, err := shellImpersonation()
	if err != nil {
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
