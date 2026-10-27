package geo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

type Place struct {
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
	Loc     string `json:"loc"`
}

func (p Place) Coords() (lat, lon float64, ok bool) {
	parts := strings.Split(p.Loc, ",")
	if len(parts) != 2 {
		return 0, 0, false
	}
	lat, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lon, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	return lat, lon, err1 == nil && err2 == nil
}

func DistanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371.0
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLon := (lon2 - lon1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * r * math.Asin(math.Sqrt(a))
}

var client = &http.Client{Timeout: 6 * time.Second}

func getJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Toque (+https://github.com/dimelim/toque)")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s respondió %s", req.URL.Host, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func Locate(ctx context.Context, ip string) (Place, error) {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return Place{}, err
	}
	var p Place
	if err := getJSON(ctx, "https://ipinfo.io/"+addr.String()+"/json", &p); err != nil {
		return Place{}, err
	}
	if p.City == "" || p.Country == "" {
		return Place{}, errors.New("ipinfo no sabe dónde está esa IP")
	}
	return p, nil
}

func Self(ctx context.Context) (Place, error) {
	var p Place
	if err := getJSON(ctx, "https://ipinfo.io/json", &p); err != nil {
		return Place{}, err
	}
	if _, _, ok := p.Coords(); !ok {
		return Place{}, errors.New("ipinfo no devolvió coordenadas")
	}
	return p, nil
}
