package app

import (
	"os"
	"path/filepath"
	"testing"
)

func useTempData(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	orig := systemDir
	systemDir = func() string { return dir }
	t.Cleanup(func() { systemDir = orig })
	return dir
}

func TestWhatsNewSkipsFreshInstallsAndShowsOnceAfterUpdate(t *testing.T) {
	dir := useTempData(t)
	fresh := &Service{returning: returningUser()}
	if fresh.WhatsNew().Show {
		t.Fatal("a fresh install should not see Novedades")
	}
	if fresh.WhatsNew().Show {
		t.Fatal("still not after the version was recorded")
	}
	os.WriteFile(filepath.Join(dir, "version.txt"), []byte("0.0.1"), 0o644)
	updated := &Service{}
	if !updated.WhatsNew().Show {
		t.Fatal("an update should show Novedades")
	}
	updated.SeenWhatsNew()
	if updated.WhatsNew().Show {
		t.Fatal("Novedades showed twice")
	}
}

func TestWhatsNewShowsForUsersBeforeVersionFile(t *testing.T) {
	dir := useTempData(t)
	os.WriteFile(filepath.Join(dir, "journal.json"), []byte("{}"), 0o644)
	s := &Service{returning: returningUser()}
	if !s.WhatsNew().Show {
		t.Fatal("someone who already used Toque should see Novedades the first time")
	}
}
