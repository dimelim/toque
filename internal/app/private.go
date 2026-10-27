package app

import (
	"fmt"
	"time"

	"toque/internal/roblox"
)

func (s *Service) PlayPrivate(link string) (PlayResult, error) {
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
	s.progress("play", "join", 0.1, "Entrando en tu servidor privado")
	sess, err := roblox.JoinIf(ctx, inst, target, 90*time.Second, nil)
	if err != nil {
		if ctx.Err() != nil {
			return PlayResult{}, ctx.Err()
		}
		return PlayResult{}, fmt.Errorf("no pude entrar en tu servidor privado: %w", err)
	}
	s.progress("play", "measure", 0.6, "Comprobando a qué distancia está el servidor")
	p := joinedProbe(ctx, sess)
	return PlayResult{First: p, Final: p, Private: true}, nil
}
