package tweaks

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const memMgmtKey = `SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management`

type PagefileInfo struct {
	Drives       []string `json:"drives"`
	OnHDD        bool     `json:"onHdd"`
	Automatic    bool     `json:"automatic"`
	SystemDrive  string   `json:"systemDrive"`
	SystemIsSSD  bool     `json:"systemIsSsd"`
	SystemFreeGB float64  `json:"systemFreeGb"`
	Configured   []string `json:"configured"`
}

func readMulti(name string) []string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, memMgmtKey, registry.QUERY_VALUE)
	if err != nil {
		return nil
	}
	defer k.Close()
	v, _, err := k.GetStringsValue(name)
	if err != nil {
		return nil
	}
	return v
}

func driveOf(path string) string {
	p := strings.TrimPrefix(path, `\??\`)
	if len(p) >= 2 && p[1] == ':' {
		return strings.ToUpper(p[:1])
	}
	return ""
}

type storagePropertyQuery struct {
	PropertyID uint32
	QueryType  uint32
	Extra      [1]byte
}

type seekPenaltyDescriptor struct {
	Version           uint32
	Size              uint32
	IncursSeekPenalty byte
}

func IsRotational(drive string) (bool, error) {
	path, _ := windows.UTF16PtrFromString(`\\.\` + drive + ":")
	h, err := windows.CreateFile(path, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		return false, err
	}
	defer windows.CloseHandle(h)
	q := storagePropertyQuery{PropertyID: 7, QueryType: 0}
	var d seekPenaltyDescriptor
	var ret uint32
	const ioctlStorageQueryProperty = 0x002D1400
	err = windows.DeviceIoControl(h, ioctlStorageQueryProperty, (*byte)(unsafe.Pointer(&q)), uint32(unsafe.Sizeof(q)), (*byte)(unsafe.Pointer(&d)), uint32(unsafe.Sizeof(d)), &ret, nil)
	if err != nil {
		return false, err
	}
	return d.IncursSeekPenalty != 0, nil
}

func freeGB(drive string) float64 {
	root, _ := windows.UTF16PtrFromString(drive + `:\`)
	var avail, total, free uint64
	if windows.GetDiskFreeSpaceEx(root, &avail, &total, &free) != nil {
		return 0
	}
	return float64(avail) / (1 << 30)
}

func Pagefile() PagefileInfo {
	info := PagefileInfo{Configured: readMulti("PagingFiles")}
	for _, p := range readMulti("ExistingPageFiles") {
		if d := driveOf(p); d != "" {
			info.Drives = append(info.Drives, d)
			if rot, err := IsRotational(d); err == nil && rot {
				info.OnHDD = true
			}
		}
	}
	for _, c := range info.Configured {
		if strings.HasPrefix(c, "?:") {
			info.Automatic = true
		}
	}
	sys := os.Getenv("SystemDrive")
	if sys == "" {
		sys = "C:"
	}
	info.SystemDrive = strings.TrimSuffix(strings.ToUpper(sys), ":")
	if rot, err := IsRotational(info.SystemDrive); err == nil {
		info.SystemIsSSD = !rot
	}
	info.SystemFreeGB = freeGB(info.SystemDrive)
	return info
}

type pagefileSSD struct{}

func (pagefileSSD) ID() string { return "pagefile-ssd" }
func (pagefileSSD) Kind() Kind { return Reboot }

func (pagefileSSD) State(Context) State {
	pf := Pagefile()
	if !pf.SystemIsSSD {
		return State{Applicable: false, Detail: "el disco del sistema no es SSD"}
	}
	onSystem := false
	for _, c := range pf.Configured {
		if driveOf(c) == pf.SystemDrive {
			onSystem = true
		}
	}
	if !pf.OnHDD && (onSystem || len(pf.Drives) > 0) {
		return State{Applicable: true, Applied: true, Detail: "archivo de paginación en " + strings.Join(pf.Drives, ", ")}
	}
	if pf.SystemFreeGB < 14 {
		return State{Applicable: false, Detail: fmt.Sprintf("solo %.0f GB libres en %s:", pf.SystemFreeGB, pf.SystemDrive)}
	}
	return State{Applicable: true, Applied: false, Detail: "archivo de paginación en disco mecánico " + strings.Join(pf.Drives, ", ")}
}

func (pagefileSSD) Snapshot(Context) (Undo, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, memMgmtKey, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()
	v, _, err := k.GetStringsValue("PagingFiles")
	if errors.Is(err, registry.ErrNotExist) {
		return Undo{"missing": "true"}, nil
	}
	if err != nil {
		return nil, err
	}
	return Undo{"prev": strings.Join(v, "\n")}, nil
}

func (pagefileSSD) Apply(Context) error {
	pf := Pagefile()
	if !pf.SystemIsSSD {
		return errors.New("system drive is not an SSD")
	}
	maxMB := 12288
	if room := int((pf.SystemFreeGB - 10) * 1024); room < maxMB {
		maxMB = room
	}
	if maxMB < 4096 {
		return errors.New("not enough free space on the system drive")
	}
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, memMgmtKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringsValue("PagingFiles", []string{fmt.Sprintf(`%s:\pagefile.sys 4096 %d`, pf.SystemDrive, maxMB)})
}

func (pagefileSSD) Revert(_ Context, u Undo) error {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, memMgmtKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if u["missing"] == "true" {
		return ignoreMissing(k.DeleteValue("PagingFiles"))
	}
	var prev []string
	if u["prev"] != "" {
		prev = strings.Split(u["prev"], "\n")
	}
	return k.SetStringsValue("PagingFiles", prev)
}

func init() {
	Register(pagefileSSD{})
}
