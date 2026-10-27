package winx

import (
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/sys/windows"
)

const dataDirSDDL = "D:PAI(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;0x1200a9;;;BU)"

var secureOnce sync.Once

func LocalAppData() string {
	if p, err := windows.KnownFolderPath(windows.FOLDERID_LocalAppData, 0); err == nil {
		return p
	}
	return os.Getenv("LOCALAPPDATA")
}

func ProgramData() string {
	if p, err := windows.KnownFolderPath(windows.FOLDERID_ProgramData, 0); err == nil {
		return p
	}
	return os.Getenv("ProgramData")
}

func secureRoot(root string) {
	if !IsElevated() {
		return
	}
	sd, err := windows.SecurityDescriptorFromString(dataDirSDDL)
	if err != nil {
		return
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return
	}
	windows.SetNamedSecurityInfo(root, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil)
}

func DataDir(parts ...string) string {
	root := filepath.Join(ProgramData(), "Toque")
	os.MkdirAll(root, 0o755)
	secureOnce.Do(func() { secureRoot(root) })
	p := filepath.Join(append([]string{root}, parts...)...)
	os.MkdirAll(p, 0o755)
	return p
}
