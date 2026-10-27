package app

import (
	"context"
	"errors"
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

type seenServer struct {
	key string
	at  time.Time
}

var (
	seenMu      sync.Mutex
	seenServers = map[string]seenServer{}
)

func noteServer(jobID, key string) {
	if jobID == "" {
		return
	}
	seenMu.Lock()
	defer seenMu.Unlock()
	seenServers[jobID] = seenServer{key: key, at: time.Now()}
}

func recentServersIn(key string, n int) []roblox.PublicServer {
	seenMu.Lock()
	defer seenMu.Unlock()
	var ids []string
	for id, sv := range seenServers {
		if sv.key == key && time.Since(sv.at) < 45*time.Minute {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return seenServers[ids[i]].at.After(seenServers[ids[j]].at) })
	var out []roblox.PublicServer
	for i := 0; i < len(ids) && i < n; i++ {
		out = append(out, roblox.PublicServer{ID: ids[i]})
	}
	return out
}

const regionTries = 8

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
	city := key[strings.Index(key, "/")+1:]
	s.progress("play", "list", 0.04, "Consultando la lista pública de servidores")
	known := recentServersIn(key, 3)
	skip := map[string]bool{}
	for _, k := range known {
		skip[k.ID] = true
	}
	servers, err := roblox.PublicServers(ctx, roblox.RealisticStreetSoccer, 1)
	var limited *roblox.RateLimitError
	if len(servers) == 0 && errors.As(err, &limited) {
		s.progress("play", "list", 0.04, fmt.Sprintf("Roblox limita las consultas a la lista de servidores. Espero %.0f segundos", limited.Wait.Seconds()))
		select {
		case <-ctx.Done():
			return PlayResult{}, ctx.Err()
		case <-time.After(limited.Wait):
		}
		servers, _ = roblox.PublicServers(ctx, roblox.RealisticStreetSoccer, 1)
	}
	candidates := append(known, roblox.PublicServer{})
	candidates = append(candidates, roblox.SampleCandidates(servers, skip, regionTries, time.Now().UnixNano())...)
	res := PlayResult{Region: key}
	seen := ""
	inRegion := func(sess roblox.Session) bool {
		p, ok := serverPlace(ctx, sess.ServerIP, sess.DatacenterID)
		if !ok {
			return false
		}
		seen = displayCity(p.City)
		noteServer(sess.JobID, regionKey(p))
		return regionKey(p) == key
	}
	for i, c := range candidates {
		if ctx.Err() != nil {
			return res, ctx.Err()
		}
		msg := fmt.Sprintf("Buscando un servidor en %s, intento %d de %d", city, i+1, len(candidates))
		if seen != "" {
			msg += ". El anterior era de " + seen
		}
		s.progress("play", "hop", 0.06+0.86*float64(i)/float64(len(candidates)), msg)
		seen = ""
		res.Tried++
		sess, err := roblox.JoinIf(ctx, inst, roblox.Target{PlaceID: roblox.RealisticStreetSoccer, JobID: c.ID}, 30*time.Second, inRegion)
		if err != nil {
			continue
		}
		res.Found = true
		res.Final = joinedProbe(ctx, sess)
		res.First = res.Final
		return res, nil
	}
	if ctx.Err() != nil {
		return res, ctx.Err()
	}
	near, err := s.playNear(ctx, inst)
	near.Region, near.Tried = key, res.Tried
	return near, err
}

func joinedProbe(ctx context.Context, sess roblox.Session) ServerProbe {
	p, _ := serverRTT(ctx, sess.ServerIP)
	p.JobID, p.Current = sess.JobID, true
	withPlace(ctx, &p, sess.DatacenterID)
	return p
}
