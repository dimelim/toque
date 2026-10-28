package winx

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	spiGetAnimation      = 0x0048
	spiSetAnimation      = 0x0049
	spifUpdateAndNotify  = 0x01 | 0x02
	wmSettingChange      = 0x001A
	hwndBroadcast        = 0xFFFF
	smtoAbortIfHung      = 0x0002
	personalizeKey       = `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`
	transparencyValue    = "EnableTransparency"
	gameDVRKey           = `Software\Microsoft\Windows\CurrentVersion\GameDVR`
	recordingValue       = "HistoricalCaptureEnabled"
	settingChangeTimeout = 1000
)

var (
	procSystemParametersInfoW = user32.NewProc("SystemParametersInfoW")
	procSendMessageTimeoutW   = user32.NewProc("SendMessageTimeoutW")
)

type visualFlag struct {
	key      string
	get, set uintptr
}

var visualFlags = []visualFlag{
	{"clientArea", 0x1042, 0x1043},
	{"menus", 0x1002, 0x1003},
	{"comboBoxes", 0x1004, 0x1005},
	{"smoothLists", 0x1006, 0x1007},
	{"selectionFade", 0x1014, 0x1015},
	{"tooltips", 0x1016, 0x1017},
	{"cursorShadow", 0x101A, 0x101B},
	{"dropShadow", 0x1024, 0x1025},
}

type VisualFX map[string]bool

type animationInfo struct {
	Size       uint32
	MinAnimate int32
}

func ReadVisualFX() (VisualFX, error) {
	fx := VisualFX{}
	for _, f := range visualFlags {
		var on int32
		if r, _, err := procSystemParametersInfoW.Call(f.get, 0, uintptr(unsafe.Pointer(&on)), 0); r == 0 {
			return nil, err
		}
		fx[f.key] = on != 0
	}
	ai := animationInfo{Size: uint32(unsafe.Sizeof(animationInfo{}))}
	if r, _, err := procSystemParametersInfoW.Call(spiGetAnimation, uintptr(ai.Size), uintptr(unsafe.Pointer(&ai)), 0); r == 0 {
		return nil, err
	}
	fx["windows"] = ai.MinAnimate != 0
	fx["transparency"] = dword(personalizeKey, transparencyValue, true)
	fx["backgroundRecording"] = dword(gameDVRKey, recordingValue, false)
	return fx, nil
}

func ApplyVisualFX(fx VisualFX) error {
	var errs []error
	for _, f := range visualFlags {
		on, ok := fx[f.key]
		if !ok {
			continue
		}
		if r, _, err := procSystemParametersInfoW.Call(f.set, 0, boolArg(on), spifUpdateAndNotify); r == 0 {
			errs = append(errs, err)
		}
	}
	if on, ok := fx["windows"]; ok {
		ai := animationInfo{Size: uint32(unsafe.Sizeof(animationInfo{}))}
		if on {
			ai.MinAnimate = 1
		}
		if r, _, err := procSystemParametersInfoW.Call(spiSetAnimation, uintptr(ai.Size), uintptr(unsafe.Pointer(&ai)), spifUpdateAndNotify); r == 0 {
			errs = append(errs, err)
		}
	}
	if on, ok := fx["transparency"]; ok && on != dword(personalizeKey, transparencyValue, true) {
		errs = append(errs, setDword(personalizeKey, transparencyValue, on))
		broadcastColorSet()
	}
	if on, ok := fx["backgroundRecording"]; ok && on != dword(gameDVRKey, recordingValue, false) {
		errs = append(errs, setDword(gameDVRKey, recordingValue, on))
	}
	return errors.Join(errs...)
}

func (fx VisualFX) Off() VisualFX {
	off := VisualFX{}
	for k := range fx {
		off[k] = false
	}
	return off
}

func (fx VisualFX) AllOff() bool {
	for _, on := range fx {
		if on {
			return false
		}
	}
	return true
}

func boolArg(on bool) uintptr {
	if on {
		return 1
	}
	return 0
}

func dword(path, name string, missing bool) bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, path, registry.QUERY_VALUE)
	if err != nil {
		return missing
	}
	defer k.Close()
	v, _, err := k.GetIntegerValue(name)
	if err != nil {
		return missing
	}
	return v != 0
}

func setDword(path, name string, on bool) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	v := uint32(0)
	if on {
		v = 1
	}
	return k.SetDWordValue(name, v)
}

func broadcastColorSet() {
	area, _ := windows.UTF16PtrFromString("ImmersiveColorSet")
	var result uintptr
	procSendMessageTimeoutW.Call(hwndBroadcast, wmSettingChange, 0, uintptr(unsafe.Pointer(area)), smtoAbortIfHung, settingChangeTimeout, uintptr(unsafe.Pointer(&result)))
}
