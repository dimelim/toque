package app

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"toque/internal/roblox"
	"toque/internal/rovalra"
)

var (
	knownMu      sync.Mutex
	knownDC      = map[string]int{}
	unknownSince = map[string]time.Time{}
)

func noteDatacenter(jobID string, datacenter int) {
	if jobID == "" || datacenter == 0 {
		return
	}
	knownMu.Lock()
	defer knownMu.Unlock()
	knownDC[jobID] = datacenter
	delete(unknownSince, jobID)
}

func serverDatacenters(ctx context.Context, servers []roblox.PublicServer) (map[string]int, error) {
	knownMu.Lock()
	var ask []string
	for _, sv := range servers {
		if _, ok := knownDC[sv.ID]; ok {
			continue
		}
		if at, ok := unknownSince[sv.ID]; ok && time.Since(at) < 5*time.Minute {
			continue
		}
		ask = append(ask, sv.ID)
	}
	knownMu.Unlock()
	var lookupErr error
	if len(ask) > 0 {
		got, err := rovalra.ServerDatacenters(ctx, roblox.RealisticStreetSoccer, ask)
		lookupErr = err
		knownMu.Lock()
		for _, id := range ask {
			if dc, ok := got[id]; ok {
				knownDC[id] = dc
				delete(unknownSince, id)
			} else if err == nil {
				unknownSince[id] = time.Now()
			}
		}
		knownMu.Unlock()
	}
	knownMu.Lock()
	defer knownMu.Unlock()
	out := make(map[string]int, len(servers))
	for _, sv := range servers {
		if dc, ok := knownDC[sv.ID]; ok {
			out[sv.ID] = dc
		}
	}
	return out, lookupErr
}

func serversIn(ctx context.Context, key string, servers []roblox.PublicServer) ([]roblox.PublicServer, error) {
	dcs, lookupErr := serverDatacenters(ctx, servers)
	var out []roblox.PublicServer
	for _, sv := range servers {
		if sv.Playing == 0 || sv.Playing >= sv.MaxPlayers {
			continue
		}
		if p, ok := datacenterPlace(ctx, dcs[sv.ID]); ok && regionKey(p) == key {
			out = append(out, sv)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		roomI, roomJ := out[i].MaxPlayers-out[i].Playing >= 2, out[j].MaxPlayers-out[j].Playing >= 2
		if roomI != roomJ {
			return roomI
		}
		return out[i].Playing > out[j].Playing
	})
	return out, lookupErr
}

func (s *Service) listPage(ctx context.Context, task string, order roblox.SortOrder, cursor string) ([]roblox.PublicServer, string, error) {
	page, next, err := roblox.PublicPage(ctx, roblox.RealisticStreetSoccer, order, cursor)
	var limited *roblox.RateLimitError
	if !errors.As(err, &limited) {
		return page, next, err
	}
	s.progress(task, "list", 0.05, fmt.Sprintf("Roblox limita las consultas a la lista de servidores. Espero %.0f segundos", limited.Wait.Seconds()))
	select {
	case <-ctx.Done():
		return nil, "", ctx.Err()
	case <-time.After(limited.Wait):
	}
	return roblox.PublicPage(ctx, roblox.RealisticStreetSoccer, order, cursor)
}

func (s *Service) joinIn(ctx context.Context, inst roblox.Install, key string, matches []roblox.PublicServer, task string) (roblox.Session, int, error) {
	accept := func(sess roblox.Session) bool {
		noteDatacenter(sess.JobID, sess.DatacenterID)
		p, ok := serverPlace(ctx, sess.ServerIP, sess.DatacenterID)
		return ok && regionKey(p) == key
	}
	tried := 0
	for _, sv := range matches {
		if tried == 4 || ctx.Err() != nil {
			break
		}
		tried++
		s.progress(task, "join", 0.2+0.15*float64(tried), fmt.Sprintf("Entrando en una partida de %s con %d jugadores", regionCity(key), sv.Playing))
		sess, err := roblox.JoinIf(ctx, inst, roblox.Target{PlaceID: roblox.RealisticStreetSoccer, JobID: sv.ID}, 30*time.Second, accept)
		if err == nil {
			return sess, tried, nil
		}
	}
	if ctx.Err() != nil {
		return roblox.Session{}, tried, ctx.Err()
	}
	return roblox.Session{}, tried, errors.New("no server in region")
}
