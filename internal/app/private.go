package app

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"toque/internal/geo"
	"toque/internal/roblox"
)

var (
	openedMu sync.Mutex
	opened   = map[string]time.Time{}
)

func openedKey(link, region string) string {
	return strings.TrimSpace(link) + "|" + region
}

func recentlyOpened(link, region string) bool {
	openedMu.Lock()
	defer openedMu.Unlock()
	at, ok := opened[openedKey(link, region)]
	return ok && time.Since(at) < 10*time.Minute
}

func rememberOpened(link, region string) {
	openedMu.Lock()
	defer openedMu.Unlock()
	opened[openedKey(link, region)] = time.Now()
}

func (s *Service) PlayPrivate(link, region string) (PlayResult, error) {
	target, err := roblox.ParsePrivateLink(link)
	if err != nil {
		return PlayResult{}, err
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
	res := PlayResult{Private: true, Region: region}
	if region != "" && !recentlyOpened(link, region) {
		s.progress("play", "vpn", 0.05, "Comprobando por qué país sale tu conexión")
		exit, err := geo.Self(ctx)
		if err != nil {
			return res, fmt.Errorf("no pude comprobar por qué país sale tu conexión: %w", err)
		}
		if country, _, _ := strings.Cut(region, "/"); exit.Country != country {
			res.NeedVPN, res.ExitCountry = true, exit.Country
			return res, nil
		}
	}
	s.progress("play", "join", 0.15, "Entrando en tu servidor privado")
	sess, err := roblox.JoinIf(ctx, inst, target, 90*time.Second, nil)
	if err != nil {
		if ctx.Err() != nil {
			return res, ctx.Err()
		}
		return res, fmt.Errorf("no pude entrar en tu servidor privado: %w", err)
	}
	s.progress("play", "measure", 0.6, "Comprobando a qué distancia está el servidor")
	res.First = joinedProbe(ctx, sess)
	res.Final = res.First
	if region != "" && res.Final.Country+"/"+res.Final.City == region {
		res.Found = true
		rememberOpened(link, region)
	}
	return res, nil
}
