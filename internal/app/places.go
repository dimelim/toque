package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"toque/internal/geo"
	"toque/internal/winx"
)

var placeMu sync.Mutex

var cityNames = map[string]string{
	"Barueri":           "São Paulo",
	"Saint-Denis":       "París",
	"Paris":             "París",
	"Frankfurt am Main": "Fráncfort",
	"New York City":     "Nueva York",
	"New York":          "Nueva York",
	"Tokyo":             "Tokio",
	"Singapore":         "Singapur",
	"Amsterdam":         "Ámsterdam",
	"London":            "Londres",
}

func displayCity(city string) string {
	if name, ok := cityNames[city]; ok {
		return name
	}
	return city
}

func regionKey(p geo.Place) string {
	return p.Country + "/" + displayCity(p.City)
}

func placesPath() string {
	return filepath.Join(winx.DataDir(), "places.json")
}

func loadPlaces() map[string]geo.Place {
	out := map[string]geo.Place{}
	if b, err := os.ReadFile(placesPath()); err == nil {
		json.Unmarshal(b, &out)
	}
	return out
}

func knownIPs() []string {
	placeMu.Lock()
	defer placeMu.Unlock()
	var out []string
	for ip := range loadPlaces() {
		out = append(out, ip)
	}
	return out
}

func placeOf(ctx context.Context, ip string) (geo.Place, bool) {
	placeMu.Lock()
	p, ok := loadPlaces()[ip]
	placeMu.Unlock()
	if ok {
		return p, true
	}
	p, err := geo.Locate(ctx, ip)
	if err != nil {
		return geo.Place{}, false
	}
	placeMu.Lock()
	defer placeMu.Unlock()
	all := loadPlaces()
	all[ip] = p
	b, _ := json.MarshalIndent(all, "", "  ")
	os.WriteFile(placesPath(), b, 0o644)
	return p, true
}

func withPlace(ctx context.Context, p *ServerProbe) {
	if p.IP == "" {
		return
	}
	if pl, ok := placeOf(ctx, p.IP); ok {
		p.City, p.Country = displayCity(pl.City), pl.Country
		noteServer(p.JobID, regionKey(pl))
	}
}
