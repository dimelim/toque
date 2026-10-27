package app

import (
	"encoding/json"
	"os"
	"path/filepath"

	"toque/internal/netq"
	"toque/internal/winx"
)

func saveLast(name string, v any) {
	if b, err := json.MarshalIndent(v, "", "  "); err == nil {
		os.WriteFile(filepath.Join(winx.DataDir(), name), b, 0o644)
	}
}

func loadLast[T any](name string) *T {
	b, err := os.ReadFile(filepath.Join(winx.DataDir(), name))
	if err != nil {
		return nil
	}
	var v T
	if json.Unmarshal(b, &v) != nil {
		return nil
	}
	return &v
}

func (s *Service) LastDiagnosis() *Diagnosis {
	return loadLast[Diagnosis]("diagnosis.json")
}

func (s *Service) LastBloat() *netq.BloatResult {
	return loadLast[netq.BloatResult]("bloat.json")
}
