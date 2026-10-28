package app

import (
	"testing"

	"toque/internal/tweaks"
)

func TestEveryPrivacyTweakHasCopy(t *testing.T) {
	for _, id := range tweaks.PrivacyIDs() {
		c, ok := windowsText[id]
		if !ok || c.title == "" || c.what == "" || (c.group != "datos" && c.group != "anuncios" && c.group != "funciones") {
			t.Fatalf("%s sin texto completo: %+v", id, c)
		}
	}
	if len(windowsText) != len(tweaks.PrivacyIDs()) {
		t.Fatalf("textos de más: %d frente a %d", len(windowsText), len(tweaks.PrivacyIDs()))
	}
}
