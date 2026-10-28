package tweaks

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"toque/internal/i18n"
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
	id     string
	values []regValue
}

const (
	absent        = "absent"
	other         = "other"
	servicePrefix = "service:"
	keyPrefix     = "key:"
)

func (v regValue) parent() string {
	root := "HKCU"
	if v.root == registry.LOCAL_MACHINE {
		root = "HKLM"
	}
	return root + `\` + v.path
}

func (v regValue) key() string {
	return v.parent() + `\` + v.name
}

func (v regValue) keyExists() bool {
	base, err := v.base()
	if err != nil {
		return true
	}
	k, err := registry.OpenKey(base, v.path, registry.QUERY_VALUE)
	if err != nil {
		return !errors.Is(err, registry.ErrNotExist)
	}
	k.Close()
	return true
}

func (v regValue) removeKeyIfEmpty() error {
	base, err := v.base()
	if err != nil {
		return err
	}
	k, err := registry.OpenKey(base, v.path, registry.QUERY_VALUE)
	if err != nil {
		return ignoreMissing(err)
	}
	st, err := k.Stat()
	k.Close()
	if err != nil || st.SubKeyCount > 0 || st.ValueCount > 0 {
		return err
	}
	return ignoreMissing(registry.DeleteKey(base, v.path))
}

func (v regValue) base() (registry.Key, error) {
	if v.root == registry.CURRENT_USER {
		return winx.UserRoot()
	}
	return v.root, nil
}

func (v regValue) read() (uint32, regState) {
	base, err := v.base()
	if err != nil {
		return 0, regOther
	}
	k, err := registry.OpenKey(base, v.path, registry.QUERY_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return 0, regAbsent
	}
	if err != nil {
		return 0, regOther
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

func (v regValue) satisfied(cur uint32, st regState) bool {
	switch {
	case st != regDword:
		return false
	case v.ceiling:
		return cur <= v.value
	}
	return cur == v.value
}

func (v regValue) write(n uint32) error {
	base, err := v.base()
	if err != nil {
		return err
	}
	k, _, err := registry.CreateKey(base, v.path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetDWordValue(v.name, n)
}

func (v regValue) remove() error {
	base, err := v.base()
	if err != nil {
		return err
	}
	k, err := registry.OpenKey(base, v.path, registry.SET_VALUE)
	if err != nil {
		return ignoreMissing(err)
	}
	defer k.Close()
	return ignoreMissing(k.DeleteValue(v.name))
}

func (p privacy) ID() string { return p.id }
func (privacy) Kind() Kind   { return Live }

func (p privacy) admin() bool {
	for _, v := range p.values {
		if v.root == registry.LOCAL_MACHINE {
			return true
		}
	}
	return false
}

func (p privacy) State(Context) State {
	st := State{Applicable: !p.admin() || winx.IsElevated(), Applied: true}
	for _, v := range p.values {
		if _, err := v.base(); err != nil {
			st.Applicable, st.Applied, st.Detail = false, false, i18n.T("No encuentro tu sesión de Windows. Vuelve a abrir Toque desde tu escritorio")
			return st
		}
		cur, rs := v.read()
		if rs == regOther {
			st.Applicable, st.Detail = false, i18n.T("Windows tiene este ajuste en un formato que Toque no toca")
		}
		if !v.satisfied(cur, rs) {
			st.Applied = false
		}
	}
	return st
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
			return nil, fmt.Errorf(i18n.T("no puedo leer %s"), v.key())
		}
		if !v.keyExists() {
			u[keyPrefix+v.parent()] = absent
		}
	}
	return u, nil
}

func (p privacy) Apply(Context) error {
	var errs []error
	for _, v := range p.values {
		if cur, st := v.read(); !v.satisfied(cur, st) && st != regOther {
			errs = append(errs, v.write(v.value))
		}
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
	for _, v := range p.values {
		if u[keyPrefix+v.parent()] == absent {
			errs = append(errs, v.removeKeyIfEmpty())
		}
	}
	for k, prev := range u {
		if name, ok := strings.CutPrefix(k, servicePrefix); ok {
			errs = append(errs, restoreService(name, prev))
		}
	}
	return errors.Join(errs...)
}

func restoreService(name, prev string) error {
	startText, delayedText, ok := strings.Cut(prev, ":")
	start, err := strconv.ParseUint(startText, 10, 32)
	if !ok || err != nil {
		return fmt.Errorf(i18n.T("estado guardado del servicio %s dañado"), name)
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
		id:     "privacy:telemetry",
		values: []regValue{{root: registry.LOCAL_MACHINE, path: `SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\DataCollection`, name: "AllowTelemetry", value: 1, ceiling: true}},
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
			{root: registry.CURRENT_USER, path: `Software\Policies\Microsoft\Windows\WindowsAI`, name: "DisableAIDataAnalysis", value: 1},
			{root: registry.CURRENT_USER, path: `Software\Policies\Microsoft\Windows\WindowsCopilot`, name: "TurnOffWindowsCopilot", value: 1},
			{root: registry.CURRENT_USER, path: explorerAdvanced, name: "ShowCopilotButton", value: 0},
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

var gameMode = regValue{root: registry.CURRENT_USER, path: `Software\Microsoft\GameBar`, name: "AutoGameModeEnabled", value: 1}

func GameMode() Tweak {
	return privacy{id: "windows:gamemode", values: []regValue{gameMode}}
}

func GameModeOff() bool {
	cur, st := gameMode.read()
	return st == regDword && cur == 0
}

func PrivacyChanges(id string) []string {
	for _, p := range privacyTweaks {
		if p.id != id {
			continue
		}
		out := make([]string, len(p.values))
		for i, v := range p.values {
			op := "="
			if v.ceiling {
				op = "≤"
			}
			out[i] = fmt.Sprintf("%s %s %d", v.key(), op, v.value)
		}
		return out
	}
	return nil
}

func PrivacyIDs() []string {
	ids := make([]string, len(privacyTweaks))
	for i, p := range privacyTweaks {
		ids[i] = p.id
	}
	return ids
}

func init() {
	for _, p := range privacyTweaks {
		Register(p)
	}
	Register(GameMode())
}
