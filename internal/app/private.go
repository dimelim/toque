package app

import (
	"errors"
	"fmt"
	"time"

	"toque/internal/geo"
	"toque/internal/i18n"
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
	inst, err := s.launchInstall()
	if err != nil {
		return PlayResult{}, err
	}
	s.progress("play", "join", 0.15, i18n.T("Entrando en tu servidor privado"))
	sess, err := roblox.JoinIf(ctx, inst, target, 90*time.Second, nil)
	if err != nil {
		if ctx.Err() != nil {
			return PlayResult{}, ctx.Err()
		}
		if errors.Is(err, roblox.ErrNotDetected) && len(roblox.Running()) > 0 {
			return PlayResult{}, errors.New(i18n.T("Roblox sigue abierto, pero no pude confirmar en qué servidor entraste"))
		}
		return PlayResult{}, fmt.Errorf(i18n.T("no pude entrar en tu servidor privado (%w)"), err)
	}
	s.progress("play", "measure", 0.6, i18n.T("Comprobando a qué distancia está el servidor"))
	p := joinedProbe(ctx, sess)
	res := PlayResult{First: p, Final: p, Private: true, Region: region}
	if region == "" {
		return res, nil
	}
	res.Found = sessionIn(ctx, sess, region)
	if res.Found {
		return res, nil
	}
	if started, ok := roblox.ServerStart(ctx, sess, 10*time.Second); ok {
		res.AlreadyOpen = started.Before(sess.LaunchedAt.Add(-10 * time.Second))
	}
	if exit, err := geo.Self(ctx); err == nil {
		res.ExitCountry = exit.Country
	}
	return res, nil
}
