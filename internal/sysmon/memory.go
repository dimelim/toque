package sysmon

import "unsafe"

var procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

func memoryStatus() memoryStatusEx {
	var ms memoryStatusEx
	ms.Length = uint32(unsafe.Sizeof(ms))
	procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&ms)))
	return ms
}

func AvailableMB() float64 {
	return float64(memoryStatus().AvailPhys) / (1 << 20)
}

func TotalMB() float64 {
	return float64(memoryStatus().TotalPhys) / (1 << 20)
}
