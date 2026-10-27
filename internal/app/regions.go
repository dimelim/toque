package app

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"toque/internal/geo"
	"toque/internal/roblox"
)

type Region struct {
	Key     string  `json:"key"`
	City    string  `json:"city"`
	Area    string  `json:"area"`
	Country string  `json:"country"`
	RTT     float64 `json:"rttMs"`
}

var (
	logOnce      sync.Once
	logSightings []roblox.Sighting
	selfOnce     sync.Once
	selfPlace    geo.Place
)

func userPlace(ctx context.Context) geo.Place {
	selfOnce.Do(func() {
		selfPlace, _ = geo.Self(ctx)
	})
	return selfPlace
}

func (s *Service) Regions() []Region {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	logOnce.Do(func() {
		if inst, err := roblox.Find(); err == nil {
			logSightings = roblox.ServerSightings(inst.LogDir())
		}
	})
	for _, si := range logSightings {
		serverPlace(ctx, si.IP, si.Datacenter)
	}
	byKey := map[string]*Region{}
	coords := map[string]geo.Place{}
	add := func(p geo.Place) *Region {
		key := regionKey(p)
		if r, ok := byKey[key]; ok {
			return r
		}
		r := &Region{Key: key, City: displayCity(p.City), Area: displayCity(p.Region), Country: p.Country}
		if r.Area == r.City {
			r.Area = ""
		}
		byKey[key], coords[key] = r, p
		return r
	}
	datacenters := robloxDatacenters(ctx)
	for _, d := range datacenters {
		add(d.Place)
	}
	rtts := freshDatacenters()
	ips := map[string]bool{}
	for _, ip := range knownIPs() {
		ips[ip] = true
	}
	for ip := range rtts {
		ips[ip] = true
	}
	for ip := range ips {
		p, ok := placeOf(ctx, ip)
		if !ok {
			continue
		}
		r, known := byKey[regionKey(p)]
		if !known {
			if len(datacenters) > 0 {
				continue
			}
			r = add(p)
		}
		if rtt, ok := rtts[ip]; ok && (r.RTT == 0 || rtt < r.RTT) {
			r.RTT = rtt
		}
	}
	self := userPlace(ctx)
	distance := func(key string) float64 {
		lat1, lon1, ok1 := self.Coords()
		lat2, lon2, ok2 := coords[key].Coords()
		if !ok1 || !ok2 {
			return math.MaxFloat64
		}
		return geo.DistanceKm(lat1, lon1, lat2, lon2)
	}
	out := make([]Region, 0, len(byKey))
	for _, r := range byKey {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if (a.RTT > 0) != (b.RTT > 0) {
			return a.RTT > 0
		}
		if a.RTT > 0 {
			return a.RTT < b.RTT
		}
		if da, db := distance(a.Key), distance(b.Key); da != db {
			return da < db
		}
		return a.City < b.City
	})
	return out
}

func regionCity(key string) string {
	_, city, _ := strings.Cut(key, "/")
	return city
}

func (s *Service) PlayIn(key string) (PlayResult, error) {
	if key == "" {
		return s.Play()
	}
	ctx, done, err := s.begin("play")
	if err != nil {
		return PlayResult{}, err
	}
	defer done()
	inst, err := roblox.Find()
	if err != nil {
		return PlayResult{}, err
	}
	res := PlayResult{Region: key}
	s.progress("play", "list", 0.05, fmt.Sprintf("Buscando partidas de %s", regionCity(key)))
	var servers []roblox.PublicServer
	var listErr error
	for _, order := range []roblox.SortOrder{roblox.FullestFirst, roblox.EmptiestFirst} {
		page, _, err := s.listPage(ctx, "play", order, "")
		if err != nil {
			if ctx.Err() != nil {
				return res, ctx.Err()
			}
			listErr = err
			continue
		}
		servers = append(servers, page...)
	}
	if len(servers) == 0 && listErr != nil {
		return res, fmt.Errorf("no pude leer la lista de servidores: %w", listErr)
	}
	matches, lookupErr := serversIn(ctx, key, servers)
	if len(matches) == 0 {
		if lookupErr != nil {
			return res, fmt.Errorf("no pude consultar RoValra: %w", lookupErr)
		}
		return res, nil
	}
	sess, tried, err := s.joinIn(ctx, inst, key, matches, "play")
	res.Tried = tried
	if err != nil {
		if ctx.Err() != nil {
			return res, ctx.Err()
		}
		return res, nil
	}
	res.Found = true
	res.Final = joinedProbe(ctx, sess)
	res.First = res.Final
	return res, nil
}
func joinedProbe(ctx context.Context, sess roblox.Session) ServerProbe {
	p, _ := serverRTT(ctx, sess.ServerIP)
	p.JobID, p.Current = sess.JobID, true
	withPlace(ctx, &p, sess.DatacenterID)
	return p
}
