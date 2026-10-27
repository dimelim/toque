package winx

import (
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Process struct {
	PID       uint32
	ParentPID uint32
	Name      string
}

func Processes() ([]Process, error) {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snap)
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	var out []Process
	for err = windows.Process32First(snap, &entry); err == nil; err = windows.Process32Next(snap, &entry) {
		out = append(out, Process{
			PID:       entry.ProcessID,
			ParentPID: entry.ParentProcessID,
			Name:      windows.UTF16ToString(entry.ExeFile[:]),
		})
	}
	return out, nil
}

func FindByName(name string) []Process {
	all, err := Processes()
	if err != nil {
		return nil
	}
	var out []Process
	for _, p := range all {
		if strings.EqualFold(p.Name, name) {
			out = append(out, p)
		}
	}
	return out
}

func Alive(pid uint32) bool {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)
	var code uint32
	if windows.GetExitCodeProcess(h, &code) != nil {
		return false
	}
	return code == 259
}

func Terminate(pid uint32) error {
	h, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, pid)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	return windows.TerminateProcess(h, 1)
}

func ImagePath(pid uint32) (string, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(h)
	buf := make([]uint16, windows.MAX_LONG_PATH)
	size := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(h, 0, &buf[0], &size); err != nil {
		return "", err
	}
	return windows.UTF16ToString(buf[:size]), nil
}

func IsElevated() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}
