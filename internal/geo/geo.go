package geo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"time"
)

type Place struct {
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
}

var client = &http.Client{Timeout: 6 * time.Second}

func Locate(ctx context.Context, ip string) (Place, error) {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return Place{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://ipinfo.io/"+addr.String()+"/json", nil)
	if err != nil {
		return Place{}, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return Place{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Place{}, fmt.Errorf("ipinfo respondió %s", resp.Status)
	}
	var p Place
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return Place{}, err
	}
	if p.City == "" || p.Country == "" {
		return Place{}, errors.New("ipinfo no sabe dónde está esa IP")
	}
	return p, nil
}
