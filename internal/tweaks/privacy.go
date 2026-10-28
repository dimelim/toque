package tweaks

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"toque/internal/winx"
)

type regValue struct {
	root    registry.Key
	path    string
	name    string
	value   uint32
	ceiling bool
}

type regState int

const (
	regAbsent regState = iota
	regDword
	regOther
)

type privacy struct {
	id      string
	values  []regValue
	service string
}

const (
	absent = "absent"
	other  = "other"
)

func (v regValue) key() string {
	root := "HKCU"
	if v.root == registry.LOCAL_MACHINE {
		root = "HKLM"
	}
	return root + `\` + v.path + `\` + v.name
}

func (v regValue) base() registry.Key {
	if v.root == registry.CURRENT_USER {
		return winx.UserRoot()
	}
	return v.root
}

func (v regValue) read() (uint32, regState) {
	k, err := registry.OpenKey(v.base(), v.path, registry.QUERY_VALUE)
	if err != nil {
		return 0, regAbsent
	}
	defer k.Close()
	n, _, err := k.GetIntegerValue(v.name)
	switch {
	case err == nil:
		return uint32(n), regDword
	case errors.Is(err, registry.ErrNotExist):
		return 0, regAbsent
	}
	return 0, regOther
}

func (v regValue) satisfied() bool {
	cur, st := v.read()
	switch {
	case st == regOther:
		return true
	case st == regAbsent:
		return false
	case v.ceiling:
		return cur <= v.value
	}
	return cur == v.value
}

func (v regValue) write(n uint32) error {
	k, _, err := registry.CreateKey(v.base(), v.path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetDWordValue(v.name, n)
}

func (v regValue) remove() error {
	k, err := registry.OpenKey(v.base(), v.path, registry.SET_VALUE)
	if err != nil {
		return ignoreMissing(err)
	}
	defer k.Close()
	return ignoreMissing(k.DeleteValue(v.name))
}

func (p privacy) ID() string { return p.id }
func (privacy) Kind() Kind   { return Live }

func (p privacy) admin() bool {
	if p.service != "" {
		return true
	}
	for _, v := range p.values {
		if v.root == registry.LOCAL_MACHINE {
			return true
		}
	}
	return false
}

func (p privacy) State(Context) State {
	applied := true
	for _, v := range p.values {
		if !v.satisfied() {
			applied = false
		}
	}
	if p.service != "" {
		start, _, err := serviceStart(p.service)
		applied = applied && err == nil && start == mgr.StartDisabled
	}
	return State{Applicable: !p.admin() || winx.IsElevated(), Applied: applied}
}

func (p privacy) Snapshot(Context) (Undo, error) {
	u := Undo{}
	for _, v := range p.values {
		switch cur, st := v.read(); st {
		case regDword:
			u[v.key()] = strconv.FormatUint(uint64(cur), 10)
		case regAbsent:
			u[v.key()] = absent
		default:
			u[v.key()] = other
		}
	}
	if p.service != "" {
		start, delayed, err := serviceStart(p.service)
		if err == nil {
			u["service:"+p.service] = fmt.Sprintf("%d:%t", start, delayed)
		}
	}
	return u, nil
}

func (p privacy) Apply(Context) error {
	var errs []error
	for _, v := range p.values {
		if !v.satisfied() {
			errs = append(errs, v.write(v.value))
		}
	}
	if p.service != "" {
		errs = append(errs, disableService(p.service))
	}
	return errors.Join(errs...)
}

func (p privacy) Revert(_ Context, u Undo) error {
	var errs []error
	for _, v := range p.values {
		prev, ok := u[v.key()]
		switch {
		case !ok, prev == other:
		case prev == absent:
			errs = append(errs, v.remove())
		default:
			n, err := strconv.ParseUint(prev, 10, 32)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			errs = append(errs, v.write(uint32(n)))
		}
	}
	if prev, ok := u["service:"+p.service]; ok && p.service != "" {
		errs = append(errs, restoreService(p.service, prev))
	}
	return errors.Join(errs...)
}

func serviceStart(name string) (uint32, bool, error) {
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return 0, false, err
	}
	defer windows.CloseServiceHandle(scm)
	n, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0, false, err
	}
	h, err := windows.OpenService(scm, n, windows.SERVICE_QUERY_CONFIG)
	if err != nil {
		return 0, false, err
	}
	s := &mgr.Service{Name: name, Handle: h}
	defer s.Close()
	c, err := s.Config()
	if err != nil {
		return 0, false, err
	}
	return c.StartType, c.DelayedAutoStart, nil
}

func disableService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return err
	}
	defer s.Close()
	c, err := s.Config()
	if err != nil {
		return err
	}
	c.StartType, c.DelayedAutoStart = mgr.StartDisabled, false
	if err := s.UpdateConfig(c); err != nil {
		return err
	}
	if st, err := s.Query(); err == nil && st.State != svc.Stopped {
		s.Control(svc.Stop)
	}
	return nil
}

func restoreService(name, prev string) error {
	startText, delayedText, ok := strings.Cut(prev, ":")
	start, err := strconv.ParseUint(startText, 10, 32)
	if !ok || err != nil {
		return fmt.Errorf("estado guardado del servicio %s dañado", name)
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return err
	}
	defer s.Close()
	c, err := s.Config()
	if err != nil {
		return err
	}
	c.StartType, c.DelayedAutoStart = uint32(start), delayedText == "true"
	if err := s.UpdateConfig(c); err != nil {
		return err
	}
	if c.StartType == mgr.StartAutomatic {
		if st, err := s.Query(); err == nil && st.State == svc.Stopped {
			s.Start()
			waitRunning(s)
		}
	}
	return nil
}

func waitRunning(s *mgr.Service) {
	for i := 0; i < 20; i++ {
		if st, err := s.Query(); err != nil || st.State == svc.Running {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

const (
	contentDelivery  = `Software\Microsoft\Windows\CurrentVersion\ContentDeliveryManager`
	inputPersonal    = `Software\Microsoft\InputPersonalization`
	explorerAdvanced = `Software\Microsoft\Windows\CurrentVersion\Explorer\Advanced`
)

var privacyTweaks = []privacy{
	{
		id:      "privacy:telemetry",
		values:  []regValue{{root: registry.LOCAL_MACHINE, path: `SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\DataCollection`, name: "AllowTelemetry", value: 1, ceiling: true}},
		service: "DiagTrack",
	},
	{
		id: "privacy:typing",
		values: []regValue{
			{root: registry.CURRENT_USER, path: `Software\Microsoft\Input\TIPC`, name: "Enabled", value: 0},
			{root: registry.CURRENT_USER, path: inputPersonal, name: "RestrictImplicitTextCollection", value: 1},
			{root: registry.CURRENT_USER, path: inputPersonal, name: "RestrictImplicitInkCollection", value: 1},
			{root: registry.CURRENT_USER, path: inputPersonal + `\TrainedDataStore`, name: "HarvestContacts", value: 0},
		},
	},
	{
		id:     "privacy:feedback",
		values: []regValue{{root: registry.CURRENT_USER, path: `Software\Microsoft\Siuf\Rules`, name: "NumberOfSIUFInPeriod", value: 0}},
	},
	{
		id: "privacy:ads",
		values: []regValue{
			{root: registry.CURRENT_USER, path: `Software\Microsoft\Windows\CurrentVersion\AdvertisingInfo`, name: "Enabled", value: 0},
			{root: registry.CURRENT_USER, path: `Software\Microsoft\Windows\CurrentVersion\Privacy`, name: "TailoredExperiencesWithDiagnosticDataEnabled", value: 0},
		},
	},
	{
		id: "privacy:suggestions",
		values: []regValue{
			{root: registry.CURRENT_USER, path: contentDelivery, name: "SubscribedContent-338388Enabled", value: 0},
			{root: registry.CURRENT_USER, path: contentDelivery, name: "SubscribedContent-338389Enabled", value: 0},
			{root: registry.CURRENT_USER, path: contentDelivery, name: "SubscribedContent-353694Enabled", value: 0},
			{root: registry.CURRENT_USER, path: contentDelivery, name: "SubscribedContent-353696Enabled", value: 0},
			{root: registry.CURRENT_USER, path: contentDelivery, name: "SubscribedContent-310093Enabled", value: 0},
			{root: registry.CURRENT_USER, path: contentDelivery, name: "SystemPaneSuggestionsEnabled", value: 0},
			{root: registry.CURRENT_USER, path: contentDelivery, name: "SoftLandingEnabled", value: 0},
			{root: registry.CURRENT_USER, path: contentDelivery, name: "SilentInstalledAppsEnabled", value: 0},
			{root: registry.CURRENT_USER, path: contentDelivery, name: "RotatingLockScreenOverlayEnabled", value: 0},
			{root: registry.CURRENT_USER, path: contentDelivery, name: "SubscribedContent-338387Enabled", value: 0},
			{root: registry.CURRENT_USER, path: explorerAdvanced, name: "Start_IrisRecommendations", value: 0},
			{root: registry.CURRENT_USER, path: explorerAdvanced, name: "ShowSyncProviderNotifications", value: 0},
			{root: registry.CURRENT_USER, path: `Software\Microsoft\Windows\CurrentVersion\UserProfileEngagement`, name: "ScoobeSystemSettingEnabled", value: 0},
		},
	},
	{
		id: "privacy:ai",
		values: []regValue{
			{root: registry.CURRENT_USER, path: `Software\Policies\Microsoft\Windows\WindowsCopilot`, name: "TurnOffWindowsCopilot", value: 1},
			{root: registry.CURRENT_USER, path: explorerAdvanced, name: "ShowCopilotButton", value: 0},
			{root: registry.CURRENT_USER, path: `Software\Policies\Microsoft\Windows\WindowsAI`, name: "DisableAIDataAnalysis", value: 1},
		},
	},
	{
		id:     "privacy:widgets",
		values: []regValue{{root: registry.LOCAL_MACHINE, path: `SOFTWARE\Policies\Microsoft\Dsh`, name: "AllowNewsAndInterests", value: 0}},
	},
	{
		id: "privacy:search",
		values: []regValue{
			{root: registry.CURRENT_USER, path: `Software\Microsoft\Windows\CurrentVersion\Search`, name: "BingSearchEnabled", value: 0},
			{root: registry.CURRENT_USER, path: `Software\Policies\Microsoft\Windows\Explorer`, name: "DisableSearchBoxSuggestions", value: 1},
		},
	},
}

func PrivacyIDs() []string {
	ids := make([]string, len(privacyTweaks))
	for i, p := range privacyTweaks {
		ids[i] = p.id
	}
	return ids
}

func PrivacyNeedsAdmin(id string) bool {
	for _, p := range privacyTweaks {
		if p.id == id {
			return p.admin()
		}
	}
	return false
}

func init() {
	for _, p := range privacyTweaks {
		Register(p)
	}
}
