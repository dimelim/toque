package winx

import (
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	ole32                = windows.NewLazySystemDLL("ole32.dll")
	procCoCreateInstance = ole32.NewProc("CoCreateInstance")

	clsidDeviceEnumerator  = windows.GUID{Data1: 0xBCDE0395, Data2: 0xE52F, Data3: 0x467C, Data4: [8]byte{0x8E, 0x3D, 0xC4, 0x57, 0x92, 0x91, 0x69, 0x2E}}
	iidDeviceEnumerator    = windows.GUID{Data1: 0xA95664D2, Data2: 0x9614, Data3: 0x4F35, Data4: [8]byte{0xA7, 0x46, 0xDE, 0x8D, 0xB6, 0x36, 0x17, 0xE6}}
	iidSessionManager2     = windows.GUID{Data1: 0x77AA99A0, Data2: 0x1BD6, Data3: 0x484F, Data4: [8]byte{0x8B, 0xC7, 0x2C, 0x65, 0x4C, 0x9A, 0x9B, 0x6F}}
	iidAudioSessionControl = windows.GUID{Data1: 0xBFB7FF88, Data2: 0x7239, Data3: 0x4FC9, Data4: [8]byte{0x8F, 0xA2, 0x07, 0xC9, 0x50, 0xBE, 0x9C, 0x6D}}
)

const (
	clsctxAll          = 0x17
	eRender            = 0
	deviceStateActive  = 1
	sessionStateActive = 1
	sFalse             = syscall.Errno(1)

	vtQueryInterface       = 0
	vtRelease              = 2
	vtEnumAudioEndpoints   = 3
	vtCollectionGetCount   = 3
	vtCollectionItem       = 4
	vtDeviceActivate       = 3
	vtGetSessionEnumerator = 5
	vtEnumeratorGetCount   = 3
	vtEnumeratorGetSession = 4
	vtSessionGetState      = 3
	vtSessionGetProcessId  = 14
)

type comObject unsafe.Pointer

func method(obj comObject, index uintptr) uintptr {
	vtbl := *(*unsafe.Pointer)(unsafe.Pointer(obj))
	return *(*uintptr)(unsafe.Add(vtbl, index*unsafe.Sizeof(uintptr(0))))
}

func release(obj comObject) {
	if obj != nil {
		syscall.SyscallN(method(obj, vtRelease), uintptr(obj))
	}
}

func AudiblePIDs() (map[uint32]bool, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := windows.CoInitializeEx(0, windows.COINIT_MULTITHREADED); err != nil && err != sFalse {
		return nil, err
	}
	defer windows.CoUninitialize()

	var enum comObject
	if hr, _, _ := procCoCreateInstance.Call(uintptr(unsafe.Pointer(&clsidDeviceEnumerator)), 0, clsctxAll, uintptr(unsafe.Pointer(&iidDeviceEnumerator)), uintptr(unsafe.Pointer(&enum))); hr != 0 {
		return nil, syscall.Errno(hr)
	}
	defer release(enum)

	var devices comObject
	if hr, _, _ := syscall.SyscallN(method(enum, vtEnumAudioEndpoints), uintptr(enum), eRender, deviceStateActive, uintptr(unsafe.Pointer(&devices))); hr != 0 {
		return nil, syscall.Errno(hr)
	}
	defer release(devices)

	var count uint32
	if hr, _, _ := syscall.SyscallN(method(devices, vtCollectionGetCount), uintptr(devices), uintptr(unsafe.Pointer(&count))); hr != 0 {
		return nil, syscall.Errno(hr)
	}
	out := map[uint32]bool{}
	for i := uint32(0); i < count; i++ {
		var device comObject
		if hr, _, _ := syscall.SyscallN(method(devices, vtCollectionItem), uintptr(devices), uintptr(i), uintptr(unsafe.Pointer(&device))); hr != 0 {
			continue
		}
		audibleOn(device, out)
		release(device)
	}
	return out, nil
}

func audibleOn(device comObject, out map[uint32]bool) {
	var manager comObject
	if hr, _, _ := syscall.SyscallN(method(device, vtDeviceActivate), uintptr(device), uintptr(unsafe.Pointer(&iidSessionManager2)), clsctxAll, 0, uintptr(unsafe.Pointer(&manager))); hr != 0 {
		return
	}
	defer release(manager)
	var sessions comObject
	if hr, _, _ := syscall.SyscallN(method(manager, vtGetSessionEnumerator), uintptr(manager), uintptr(unsafe.Pointer(&sessions))); hr != 0 {
		return
	}
	defer release(sessions)
	var n int32
	if hr, _, _ := syscall.SyscallN(method(sessions, vtEnumeratorGetCount), uintptr(sessions), uintptr(unsafe.Pointer(&n))); hr != 0 {
		return
	}
	for i := int32(0); i < n; i++ {
		var control comObject
		if hr, _, _ := syscall.SyscallN(method(sessions, vtEnumeratorGetSession), uintptr(sessions), uintptr(i), uintptr(unsafe.Pointer(&control))); hr != 0 {
			continue
		}
		if pid, ok := audiblePID(control); ok {
			out[pid] = true
		}
		release(control)
	}
}

func audiblePID(control comObject) (uint32, bool) {
	var control2 comObject
	if hr, _, _ := syscall.SyscallN(method(control, vtQueryInterface), uintptr(control), uintptr(unsafe.Pointer(&iidAudioSessionControl)), uintptr(unsafe.Pointer(&control2))); hr != 0 {
		return 0, false
	}
	defer release(control2)
	var state int32
	if hr, _, _ := syscall.SyscallN(method(control2, vtSessionGetState), uintptr(control2), uintptr(unsafe.Pointer(&state))); hr != 0 || state != sessionStateActive {
		return 0, false
	}
	var pid uint32
	if hr, _, _ := syscall.SyscallN(method(control2, vtSessionGetProcessId), uintptr(control2), uintptr(unsafe.Pointer(&pid))); hr != 0 || pid == 0 {
		return 0, false
	}
	return pid, true
}
