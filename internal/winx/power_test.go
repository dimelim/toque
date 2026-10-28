package winx

import (
	"os"
	"testing"
)

func TestPowerOverlayRoundTrip(t *testing.T) {
	laptop, plugged := Laptop()
	before, err := PowerOverlay()
	t.Logf("laptop=%v plugged=%v overlay=%v err=%v", laptop, plugged, before, err)
	if err != nil {
		t.Skip("este equipo no expone el modo de energía")
	}
	if os.Getenv("TOQUE_SYSTEM_TESTS") == "" {
		t.Skip("cambia el modo de energía; actívalo con TOQUE_SYSTEM_TESTS=1")
	}
	defer func() {
		if err := SetPowerOverlay(before); err != nil {
			t.Errorf("restore: %v", err)
		}
		if after, _ := PowerOverlay(); after != before {
			t.Errorf("not restored: %v", after)
		}
	}()
	if err := SetPowerOverlay(BestPerformance); err != nil {
		t.Fatal(err)
	}
	if during, _ := PowerOverlay(); during != BestPerformance {
		t.Fatalf("overlay during = %v", during)
	}
}
