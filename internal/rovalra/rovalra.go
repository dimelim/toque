package rovalra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"toque/internal/geo"
	"toque/internal/i18n"
)

const base = "https://apis.rovalra.com"

var client = &http.Client{Timeout: 8 * time.Second}

func get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
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
		return fmt.Errorf(i18n.T("RoValra respondió %s"), resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

type Datacenter struct {
	IDs   []int     `json:"ids"`
	Place geo.Place `json:"place"`
}

type location struct {
	DataCenterIDs []int `json:"dataCenterIds"`
	Location      struct {
		City    string   `json:"city"`
		Region  string   `json:"region"`
		Country string   `json:"country"`
		LatLong []string `json:"latLong"`
	} `json:"location"`
	Inactive bool `json:"inactive"`
}

func Datacenters(ctx context.Context) ([]Datacenter, error) {
	var raw []location
	if err := get(ctx, "/v1/datacenters/list", &raw); err != nil {
		return nil, err
	}
	var out []Datacenter
	for _, l := range raw {
		if l.Inactive || l.Location.City == "" || len(l.DataCenterIDs) == 0 {
			continue
		}
		p := geo.Place{City: l.Location.City, Region: l.Location.Region, Country: l.Location.Country}
		if len(l.Location.LatLong) == 2 {
			p.Loc = l.Location.LatLong[0] + "," + l.Location.LatLong[1]
		}
		out = append(out, Datacenter{IDs: l.DataCenterIDs, Place: p})
	}
	if len(out) == 0 {
		return nil, errors.New(i18n.T("RoValra devolvió una lista vacía"))
	}
	return out, nil
}

const serversPerRequest = 50

func ServerDatacenters(ctx context.Context, placeID int64, ids []string) (map[string]int, error) {
	out := map[string]int{}
	for start := 0; start < len(ids); start += serversPerRequest {
		end := min(start+serversPerRequest, len(ids))
		var page struct {
			Servers []struct {
				ServerID     string `json:"server_id"`
				DatacenterID int    `json:"datacenter_id"`
			} `json:"servers"`
		}
		path := fmt.Sprintf("/v1/servers/details?place_id=%d&server_ids=%s", placeID, strings.Join(ids[start:end], ","))
		if err := get(ctx, path, &page); err != nil {
			return out, err
		}
		for _, s := range page.Servers {
			if s.DatacenterID > 0 {
				out[s.ServerID] = s.DatacenterID
			}
		}
	}
	return out, nil
}
