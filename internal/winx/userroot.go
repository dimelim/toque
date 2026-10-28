package winx

import (
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

var (
	procRegOpenCurrentUser = advapi32.NewProc("RegOpenCurrentUser")
	userRootMu             sync.Mutex
	userRootKey            registry.Key
)

func UserRoot() (registry.Key, error) {
	if !IsElevated() {
		return registry.CURRENT_USER, nil
	}
	userRootMu.Lock()
	defer userRootMu.Unlock()
	if userRootKey != 0 {
		return userRootKey, nil
	}
	var h syscall.Handle
	err := AsUser(func() error {
		if r, _, _ := procRegOpenCurrentUser.Call(uintptr(registry.ALL_ACCESS), uintptr(unsafe.Pointer(&h))); r != 0 {
			return syscall.Errno(r)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	userRootKey = registry.Key(h)
	return userRootKey, nil
}
