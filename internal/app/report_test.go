package app

import (
	"archive/zip"
	"path/filepath"
	"testing"
)

func TestWriteReportCreatesZipWithInfo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "informe.zip")
	if err := writeReport(path); err != nil {
		t.Fatal(err)
	}
	r, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	found := false
	for _, f := range r.File {
		t.Log(f.Name, f.UncompressedSize64)
		found = found || f.Name == "info.json"
	}
	if !found {
		t.Fatal("info.json missing from report")
	}
}
