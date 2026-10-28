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
	"toque/internal/i18n"
	"toque/internal/winx"
)

var (
	placeMu  sync.Mutex
	placeMem = map[string]geo.Place{}
)

var cityNames = map[string]string{
	"Barueri":               "São Paulo",
	"Saint-Denis":           i18n.Mark("París"),
	"Paris":                 i18n.Mark("París"),
	"Leesburg":              "Ashburn",
	"Frankfurt am Main":     i18n.Mark("Fráncfort"),
	"New York City":         i18n.Mark("Nueva York"),
	"New York":              i18n.Mark("Nueva York"),
	"Los Angeles":           i18n.Mark("Los Ángeles"),
	"Santiago de Querétaro": i18n.Mark("Querétaro"),
	"Tokyo":                 i18n.Mark("Tokio"),
	"Singapore":             i18n.Mark("Singapur"),
	"Amsterdam":             i18n.Mark("Ámsterdam"),
	"London":                i18n.Mark("Londres"),
	"Zürich":                i18n.Mark("Zúrich"),
	"Warsaw":                i18n.Mark("Varsovia"),
	"Istanbul":              i18n.Mark("Estambul"),
	"Cape Town":             i18n.Mark("Ciudad del Cabo"),
	"Athens":                i18n.Mark("Atenas"),
	"Milan":                 i18n.Mark("Milán"),
	"Dublin":                i18n.Mark("Dublín"),
	"Jakarta":               i18n.Mark("Yakarta"),
	"Mumbai":                i18n.Mark("Bombay"),
	"Sydney":                i18n.Mark("Sídney"),
}

func displayCity(city string) string {
	if name, ok := cityNames[city]; ok {
		return name
	}
	return city
}

func localCity(city string) string {
	return i18n.T(displayCity(city))
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
		p.City, p.Country = localCity(pl.City), pl.Country
		noteDatacenter(p.JobID, datacenter)
	}
}
