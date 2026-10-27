package app

import (
	"fmt"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"toque/internal/sysmon"
)

type SystemInfo struct {
	CPU       string  `json:"cpu"`
	Threads   int     `json:"threads"`
	GPU       string  `json:"gpu"`
	RAMGB     float64 `json:"ramGb"`
	RefreshHz int     `json:"refreshHz"`
	OS        string  `json:"os"`
}

func regString(root registry.Key, path, name string) string {
	k, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()
	v, _, _ := k.GetStringValue(name)
	return strings.TrimSpace(v)
}

func gpuName() string {
	const class = `SYSTEM\CurrentControlSet\Control\Class\{4d36e968-e325-11ce-bfc1-08002be10318}`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, class, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return ""
	}
	defer k.Close()
	names, _ := k.ReadSubKeyNames(-1)
	var best string
	for _, n := range names {
		desc := regString(registry.LOCAL_MACHINE, class+`\`+n, "DriverDesc")
		if desc == "" || strings.Contains(desc, "Microsoft Basic") || strings.Contains(desc, "Remote") {
			continue
		}
		if best == "" || strings.Contains(desc, "NVIDIA") || strings.Contains(desc, "Radeon") {
			best = desc
		}
	}
	return best
}

var procEnumDisplaySettings = windows.NewLazySystemDLL("user32.dll").NewProc("EnumDisplaySettingsW")

type devmode struct {
	DeviceName         [32]uint16
	SpecVersion        uint16
	DriverVersion      uint16
	Size               uint16
	DriverExtra        uint16
	Fields             uint32
	Position           [2]int32
	DisplayOrientation uint32
	DisplayFixedOutput uint32
	Color              int16
	Duplex             int16
	YResolution        int16
	TTOption           int16
	Collate            int16
	FormName           [32]uint16
	LogPixels          uint16
	BitsPerPel         uint32
	PelsWidth          uint32
	PelsHeight         uint32
	DisplayFlags       uint32
	DisplayFrequency   uint32
	Rest               [8]uint32
}

func refreshRate() int {
	var dm devmode
	dm.Size = uint16(unsafe.Sizeof(dm))
	r, _, _ := procEnumDisplaySettings.Call(0, uintptr(0xFFFFFFFF), uintptr(unsafe.Pointer(&dm)))
	if r == 0 || dm.DisplayFrequency <= 1 {
		return 60
	}
	return int(dm.DisplayFrequency)
}

func ReadSystem() SystemInfo {
	cpu := regString(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\CentralProcessor\0`, "ProcessorNameString")
	cv := `SOFTWARE\Microsoft\Windows NT\CurrentVersion`
	build := regString(registry.LOCAL_MACHINE, cv, "CurrentBuild")
	product := regString(registry.LOCAL_MACHINE, cv, "ProductName")
	if b := build; b != "" {
		var n int
		fmt.Sscan(b, &n)
		if n >= 22000 {
			product = strings.Replace(product, "Windows 10", "Windows 11", 1)
		}
	}
	return SystemInfo{
		CPU:       cpu,
		Threads:   runtime.NumCPU(),
		GPU:       gpuName(),
		RAMGB:     float64(int(sysmon.TotalMB()/1024*10+0.5)) / 10,
		RefreshHz: refreshRate(),
		OS:        strings.TrimSpace(product + " " + build),
	}
}
