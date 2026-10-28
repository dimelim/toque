package app

import "testing"

func TestCloseAppMapsLabelsAndProtectsSystem(t *testing.T) {
	steam := appExes("Steam")
	if !steam["steam.exe"] || !steam["steamwebhelper.exe"] {
		t.Fatalf("Steam exes = %v", steam)
	}
	if got := appExes("notepad"); !got["notepad.exe"] || len(got) != 1 {
		t.Fatalf("notepad exes = %v", got)
	}
	s := &Service{}
	for _, name := range []string{"explorer", "MsMpEng", "toque", "exitlag", "svchost", "dwm", "RobloxPlayerBeta"} {
		if _, err := s.CloseApp(name); err == nil || err.Error() != "Toque no cierra ese programa" {
			t.Errorf("CloseApp(%q) = %v, want refusal", name, err)
		}
	}
}
