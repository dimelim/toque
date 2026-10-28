package winx

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	powrprof                           = windows.NewLazySystemDLL("powrprof.dll")
	procPowerGetEffectiveOverlayScheme = powrprof.NewProc("PowerGetEffectiveOverlayScheme")
	procPowerSetActiveOverlayScheme    = powrprof.NewProc("PowerSetActiveOverlayScheme")
	procGetSystemPowerStatus           = kernel32.NewProc("GetSystemPowerStatus")

	BestPerformance = windows.GUID{Data1: 0xded574b5, Data2: 0x45a0, Data3: 0x4f42, Data4: [8]byte{0x87, 0x37, 0x46, 0x34, 0x5c, 0x09, 0xc2, 0x38}}
)

type powerStatus struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	SystemStatusFlag    byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

func PowerOverlay() (windows.GUID, error) {
	var g windows.GUID
	if r, _, _ := procPowerGetEffectiveOverlayScheme.Call(uintptr(unsafe.Pointer(&g))); r != 0 {
		return g, syscall.Errno(r)
	}
	return g, nil
}

func SetPowerOverlay(g windows.GUID) error {
	if r, _, _ := procPowerSetActiveOverlayScheme.Call(uintptr(unsafe.Pointer(&g))); r != 0 {
		return syscall.Errno(r)
	}
	return nil
}

func Laptop() (laptop, plugged bool) {
	var st powerStatus
	if r, _, _ := procGetSystemPowerStatus.Call(uintptr(unsafe.Pointer(&st))); r == 0 {
		return false, false
	}
	return st.BatteryFlag != 128 && st.BatteryFlag != 255, st.ACLineStatus == 1
}
