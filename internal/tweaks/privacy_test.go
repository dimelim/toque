package tweaks

import (
	"testing"

	"golang.org/x/sys/windows/registry"
)

const testPath = `Software\ToqueTest`

func TestPrivacyRestoresPreviousAndAbsentValues(t *testing.T) {
	registry.DeleteKey(registry.CURRENT_USER, testPath)
	t.Cleanup(func() { registry.DeleteKey(registry.CURRENT_USER, testPath) })
	existing := regValue{registry.CURRENT_USER, testPath, "Existing", 0}
	missing := regValue{registry.CURRENT_USER, testPath, "Missing", 1}
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
	if v, ok := existing.read(); !ok || v != 7 {
		t.Fatalf("existing = %d, %v; want 7", v, ok)
	}
	if _, ok := missing.read(); ok {
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
