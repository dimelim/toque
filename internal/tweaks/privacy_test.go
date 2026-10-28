package tweaks

import (
	"testing"

	"golang.org/x/sys/windows/registry"
)

const testPath = `Software\ToqueTest`

func TestPrivacyRestoresPreviousAndAbsentValues(t *testing.T) {
	registry.DeleteKey(registry.CURRENT_USER, testPath)
	t.Cleanup(func() { registry.DeleteKey(registry.CURRENT_USER, testPath) })
	existing := regValue{root: registry.CURRENT_USER, path: testPath, name: "Existing", value: 0}
	missing := regValue{root: registry.CURRENT_USER, path: testPath, name: "Missing", value: 1}
	if err := existing.write(7); err != nil {
		t.Fatal(err)
	}
	p := privacy{id: "privacy:test", values: []regValue{existing, missing}}
	if p.State(Context{}).Applied {
		t.Fatal("applied before apply")
	}
	u, err := p.Snapshot(Context{})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Apply(Context{}); err != nil {
		t.Fatal(err)
	}
	if !p.State(Context{}).Applied {
		t.Fatal("not applied after apply")
	}
	if err := p.Revert(Context{}, u); err != nil {
		t.Fatal(err)
	}
	if v, st := existing.read(); st != regDword || v != 7 {
		t.Fatalf("existing = %d, %v; want 7", v, st)
	}
	if _, st := missing.read(); st != regAbsent {
		t.Fatal("missing value was left behind")
	}
}

func TestPrivacyTweaksAreRegistered(t *testing.T) {
	for _, id := range PrivacyIDs() {
		if _, err := Get(id); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
	}
	start, _, err := serviceStart("DiagTrack")
	t.Logf("DiagTrack start=%d err=%v", start, err)
}

func TestPrivacyNeverRaisesACeilingOrTouchesOtherTypes(t *testing.T) {
	registry.DeleteKey(registry.CURRENT_USER, testPath)
	t.Cleanup(func() { registry.DeleteKey(registry.CURRENT_USER, testPath) })
	k, _, err := registry.CreateKey(registry.CURRENT_USER, testPath, registry.SET_VALUE)
	if err != nil {
		t.Fatal(err)
	}
	k.SetDWordValue("Telemetry", 0)
	k.SetStringValue("Text", "keep")
	k.Close()
	ceiling := regValue{root: registry.CURRENT_USER, path: testPath, name: "Telemetry", value: 1, ceiling: true}
	text := regValue{root: registry.CURRENT_USER, path: testPath, name: "Text", value: 0}
	p := privacy{id: "privacy:test", values: []regValue{ceiling, text}}
	if !p.State(Context{}).Applied {
		t.Fatal("a lower value and a non-dword value should count as applied")
	}
	u, _ := p.Snapshot(Context{})
	if err := p.Apply(Context{}); err != nil {
		t.Fatal(err)
	}
	if err := p.Revert(Context{}, u); err != nil {
		t.Fatal(err)
	}
	if v, st := ceiling.read(); st != regDword || v != 0 {
		t.Fatalf("telemetry = %d, %v; want 0", v, st)
	}
	k, _ = registry.OpenKey(registry.CURRENT_USER, testPath, registry.QUERY_VALUE)
	defer k.Close()
	if s, _, err := k.GetStringValue("Text"); err != nil || s != "keep" {
		t.Fatalf("string value changed: %q %v", s, err)
	}
}
