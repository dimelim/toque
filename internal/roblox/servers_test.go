package roblox

import "testing"

func TestSampleCandidatesSkipsFullEmptyAndExcluded(t *testing.T) {
	servers := []PublicServer{
		{ID: "full", Playing: 8, MaxPlayers: 8},
		{ID: "empty", Playing: 0, MaxPlayers: 8},
		{ID: "skip", Playing: 5, MaxPlayers: 8},
		{ID: "one-slot", Playing: 7, MaxPlayers: 8},
		{ID: "two", Playing: 2, MaxPlayers: 8},
	}
	got := SampleCandidates(servers, map[string]bool{"skip": true}, 5, 1)
	ids := map[string]bool{}
	for _, s := range got {
		ids[s.ID] = true
	}
	if len(got) != 2 || !ids["one-slot"] || !ids["two"] {
		t.Fatalf("SampleCandidates = %+v, want one-slot and two", got)
	}
}
