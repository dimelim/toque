package app

import (
	"fmt"
	"time"

	"toque/internal/geo"
	"toque/internal/roblox"
)

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
	s.progress("play", "join", 0.15, "Entrando en tu servidor privado")
	sess, err := roblox.JoinIf(ctx, inst, target, 90*time.Second, nil)
	if err != nil {
		if ctx.Err() != nil {
			return PlayResult{}, ctx.Err()
		}
		return PlayResult{}, fmt.Errorf("no pude entrar en tu servidor privado: %w", err)
	}
	s.progress("play", "measure", 0.6, "Comprobando a qué distancia está el servidor")
	p := joinedProbe(ctx, sess)
	res := PlayResult{First: p, Final: p, Private: true, Region: region}
	if region == "" {
		return res, nil
	}
	res.Found = sessionIn(ctx, sess, region)
	if !res.Found {
		if exit, err := geo.Self(ctx); err == nil {
			res.ExitCountry = exit.Country
		}
	}
	return res, nil
}
