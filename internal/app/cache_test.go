package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmptyDirKeepsTheFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sounds")
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "a.ogg"), make([]byte, 1000), 0o644)
	os.WriteFile(filepath.Join(dir, "sub", "b.ogg"), make([]byte, 500), 0o644)
	freed, err := emptyDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if freed != 1500 {
		t.Fatalf("freed = %d, want 1500", freed)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("the folder should stay and be empty, got %v entries, err %v", len(entries), err)
	}
	if freed, err := emptyDir(filepath.Join(dir, "missing")); freed != 0 || err != nil {
		t.Fatalf("a missing folder is nothing to clear, got %d, %v", freed, err)
	}
}
