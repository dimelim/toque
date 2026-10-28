package winx

import (
	"os"
	"reflect"
	"testing"
)

func TestVisualFXReapplyKeepsSettings(t *testing.T) {
	before, err := ReadVisualFX()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("now: %v", before)
	if err := ApplyVisualFX(before); err != nil {
		t.Fatal(err)
	}
	after, err := ReadVisualFX()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("reapplying changed settings: before %v after %v", before, after)
	}
}

func TestVisualFXOffAndRestore(t *testing.T) {
	if os.Getenv("TOQUE_SYSTEM_TESTS") == "" {
		t.Skip("cambia los efectos visuales de Windows; actívalo con TOQUE_SYSTEM_TESTS=1")
	}
	before, err := ReadVisualFX()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := ApplyVisualFX(before); err != nil {
			t.Errorf("restore: %v", err)
		}
		if after, _ := ReadVisualFX(); !reflect.DeepEqual(before, after) {
			t.Errorf("not restored: before %v after %v", before, after)
		}
	}()
	if err := ApplyVisualFX(before.Off()); err != nil {
		t.Fatal(err)
	}
	during, err := ReadVisualFX()
	if err != nil {
		t.Fatal(err)
	}
	if !during.AllOff() {
		t.Fatalf("still on: %v", during)
	}
}
