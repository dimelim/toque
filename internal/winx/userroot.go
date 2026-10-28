package winx

import (
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

var (
	procRegOpenCurrentUser = advapi32.NewProc("RegOpenCurrentUser")
	userRootOnce           sync.Once
	userRoot               = registry.CURRENT_USER
)

func UserRoot() registry.Key {
	userRootOnce.Do(func() {
		if !IsElevated() {
			return
		}
		var h syscall.Handle
		err := AsUser(func() error {
			if r, _, _ := procRegOpenCurrentUser.Call(uintptr(registry.ALL_ACCESS), uintptr(unsafe.Pointer(&h))); r != 0 {
				return syscall.Errno(r)
			}
			return nil
		})
		if err == nil && h != 0 {
			userRoot = registry.Key(h)
		}
	})
	return userRoot
}
