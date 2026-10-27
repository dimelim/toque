package app

import (
	"context"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"toque/internal/geo"
	"toque/internal/winx"
)

var (
	placeMu  sync.Mutex
	placeMem = map[string]geo.Place{}
)

var cityNames = map[string]string{
	"Barueri":               "São Paulo",
	"Saint-Denis":           "París",
	"Paris":                 "París",
	"Leesburg":              "Ashburn",
	"Frankfurt am Main":     "Fráncfort",
	"New York City":         "Nueva York",
	"New York":              "Nueva York",
	"Los Angeles":           "Los Ángeles",
	"Santiago de Querétaro": "Querétaro",
	"Tokyo":                 "Tokio",
	"Singapore":             "Singapur",
	"Amsterdam":             "Ámsterdam",
	"London":                "Londres",
	"Zürich":                "Zúrich",
	"Warsaw":                "Varsovia",
	"Istanbul":              "Estambul",
	"Cape Town":             "Ciudad del Cabo",
	"Athens":                "Atenas",
	"Milan":                 "Milán",
	"Dublin":                "Dublín",
	"Jakarta":               "Yakarta",
	"Mumbai":                "Bombay",
	"Sydney":                "Sídney",
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
	maps.Copy(out, placeMem)
	return out
}

func rememberPlace(ip string, p geo.Place) {
	placeMu.Lock()
	defer placeMu.Unlock()
	placeMem[ip] = p
	all := loadPlaces()
	b, _ := json.MarshalIndent(all, "", "  ")
	os.WriteFile(placesPath(), b, 0o644)
}

func knownIPs() []string {
	placeMu.Lock()
	defer placeMu.Unlock()
	return slices.Collect(maps.Keys(loadPlaces()))
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
	rememberPlace(ip, p)
	return p, true
}

func serverPlace(ctx context.Context, ip string, datacenter int) (geo.Place, bool) {
	if p, ok := datacenterPlace(ctx, datacenter); ok {
		if ip != "" {
			placeMu.Lock()
			known, seen := loadPlaces()[ip]
			placeMu.Unlock()
			if !seen || known != p {
				rememberPlace(ip, p)
			}
		}
		return p, true
	}
	if ip == "" {
		return geo.Place{}, false
	}
	return placeOf(ctx, ip)
}

func withPlace(ctx context.Context, p *ServerProbe, datacenter int) {
	if pl, ok := serverPlace(ctx, p.IP, datacenter); ok {
		p.City, p.Country = displayCity(pl.City), pl.Country
		noteDatacenter(p.JobID, datacenter)
	}
}
