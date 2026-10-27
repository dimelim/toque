package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"toque/internal/geo"
	"toque/internal/winx"
)

type dcMap struct {
	FetchedAt time.Time        `json:"fetchedAt"`
	Locations []geo.Datacenter `json:"locations"`
}

var (
	dcMapMu  sync.Mutex
	dcMapMem dcMap
)

func dcMapPath() string {
	return filepath.Join(winx.DataDir(), "dcmap.json")
}

func robloxDatacenters(ctx context.Context) []geo.Datacenter {
	dcMapMu.Lock()
	defer dcMapMu.Unlock()
	if time.Since(dcMapMem.FetchedAt) < 24*time.Hour {
		return dcMapMem.Locations
	}
	var disk dcMap
	if b, err := os.ReadFile(dcMapPath()); err == nil {
		json.Unmarshal(b, &disk)
	}
	if time.Since(disk.FetchedAt) < 24*time.Hour {
		dcMapMem = disk
		return disk.Locations
	}
	fresh, err := geo.RobloxDatacenters(ctx)
	if err != nil {
		stale := dcMapMem.Locations
		if len(disk.Locations) > 0 {
			stale = disk.Locations
		}
		dcMapMem = dcMap{FetchedAt: time.Now().Add(-23 * time.Hour), Locations: stale}
		return stale
	}
	dcMapMem = dcMap{FetchedAt: time.Now(), Locations: fresh}
	b, _ := json.MarshalIndent(dcMapMem, "", "  ")
	os.WriteFile(dcMapPath(), b, 0o644)
	return fresh
}

func datacenterPlace(ctx context.Context, id int) (geo.Place, bool) {
	if id == 0 {
		return geo.Place{}, false
	}
	for _, d := range robloxDatacenters(ctx) {
		if slices.Contains(d.IDs, id) {
			return d.Place, true
		}
	}
	return geo.Place{}, false
}
