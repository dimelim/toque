package tweaks

import "testing"

func TestFlagIDRoundTripsAnyValue(t *testing.T) {
	value := "https://127.0.0.1:443/|a,b=c d+e"
	flags, err := parseFlagID(FlagID("DFStringAnalyticsNS1BeaconConfig", value))
	if err != nil || flags["DFStringAnalyticsNS1BeaconConfig"] != value {
		t.Fatalf("parseFlagID = %v, %v; want %q", flags, err, value)
	}
}

func TestParseFlagIDReadsCatalogIDs(t *testing.T) {
	flags, err := parseFlagID("flag:FIntFRMMinGrassDistance=0,FIntFRMMaxGrassDistance=0")
	if err != nil || len(flags) != 2 || flags["FIntFRMMaxGrassDistance"] != "0" {
		t.Fatalf("parseFlagID = %v, %v", flags, err)
	}
	if _, err := parseFlagID("flag:bad name=1"); err == nil {
		t.Fatal("parseFlagID accepted a name with a space")
	}
}

func TestReleaseFlagsKeepsTheUntouchedHalf(t *testing.T) {
	grass := "flag:FIntFRMMinGrassDistance=0,FIntFRMMaxGrassDistance=0"
	j := journal{Applied: map[string]Undo{grass: {"FIntFRMMinGrassDistance": "", "FIntFRMMaxGrassDistance": ""}}}
	cur := map[string]any{"FIntFRMMinGrassDistance": "0", "FIntFRMMaxGrassDistance": "0"}
	releaseFlags(j, cur, map[string]bool{"FIntFRMMinGrassDistance": true})
	if _, ok := cur["FIntFRMMinGrassDistance"]; ok {
		t.Fatalf("released flag still set: %v", cur)
	}
	if cur["FIntFRMMaxGrassDistance"] != "0" {
		t.Fatalf("untouched flag lost: %v", cur)
	}
	if _, ok := j.Applied[grass]; ok {
		t.Fatal("old two-flag entry still in the journal")
	}
	undo, ok := j.Applied[FlagID("FIntFRMMaxGrassDistance", "0")]
	if !ok || undo["FIntFRMMaxGrassDistance"] != "" {
		t.Fatalf("untouched flag not kept as its own entry: %v", j.Applied)
	}
	if !FlagsManaged(map[string]Undo{FlagID("FIntFRMMinGrassDistance", "0"): {}, FlagID("FIntFRMMaxGrassDistance", "0"): {}}, grass) {
		t.Fatal("two single entries do not count as the catalog entry")
	}
}
