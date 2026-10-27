package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

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
	logOnce sync.Once
	logIPs  []string
)

func (s *Service) Regions() []Region {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	logOnce.Do(func() {
		if inst, err := roblox.Find(); err == nil {
			logIPs = roblox.ServerIPs(inst.LogDir())
		}
	})
	rtts := freshDatacenters()
	ips := map[string]bool{}
	for _, ip := range append(knownIPs(), logIPs...) {
		ips[ip] = true
	}
	for ip := range rtts {
		ips[ip] = true
	}
	byKey := map[string]*Region{}
	for ip := range ips {
		r := Region{Key: "ip/" + ip, City: ip}
		if p, ok := placeOf(ctx, ip); ok {
			r = Region{Key: regionKey(p), City: displayCity(p.City), Area: displayCity(p.Region), Country: p.Country}
			if r.Area == r.City {
				r.Area = ""
			}
		}
		cur, ok := byKey[r.Key]
		if !ok {
			cur = &r
			byKey[r.Key] = cur
		}
		if rtt, ok := rtts[ip]; ok && (cur.RTT == 0 || rtt < cur.RTT) {
			cur.RTT = rtt
		}
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
		if a.RTT != b.RTT {
			return a.RTT < b.RTT
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
	servers, _ := roblox.PublicServers(ctx, roblox.RealisticStreetSoccer, 2)
	candidates := append(known, roblox.PublicServer{})
	candidates = append(candidates, roblox.SampleCandidates(servers, skip, regionTries, time.Now().UnixNano())...)
	res := PlayResult{Region: key}
	seen := ""
	inRegion := func(sess roblox.Session) bool {
		p, ok := placeOf(ctx, sess.ServerIP)
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
		sess, err := roblox.JoinIf(ctx, inst, roblox.RealisticStreetSoccer, c.ID, 30*time.Second, inRegion)
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
	withPlace(ctx, &p)
	return p
}
