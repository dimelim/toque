package app

import (
	"errors"
	"net/netip"
	"time"

	"toque/internal/bench"
	"toque/internal/netq"
	"toque/internal/roblox"
)

func (s *Service) Bufferbloat() (netq.BloatResult, error) {
	ctx, done, err := s.begin("bloat")
	if err != nil {
		return netq.BloatResult{}, err
	}
	defer done()
	target := "128.116.127.33"
	if inst, err := roblox.Find(); err == nil {
		if ip := lastServerIP(inst); ip != "" {
			target = ip
		}
	}
	s.progress("bloat", "path", 0.05, "Buscando el primer salto de tu proveedor")
	var dst netip.Addr
	for _, h := range netq.Trace(netip.MustParseAddr(target), 8, 2) {
		if !h.Private {
			dst = netip.MustParseAddr(h.Addr)
			break
		}
	}
	if !dst.IsValid() {
		return netq.BloatResult{}, errors.New("no encontré un salto público que responda al ping")
	}
	step := 0
	res := netq.Bufferbloat(ctx, dst, 12*time.Second, func(kind string) func() {
		step++
		msg := "Midiendo con la descarga saturada"
		if kind == "up" {
			msg = "Midiendo con la subida saturada"
		}
		s.progress("bloat", kind, 0.2+0.35*float64(step), msg)
		return bench.StartStress(kind)
	})
	return res, nil
}
