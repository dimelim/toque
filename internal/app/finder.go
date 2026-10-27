package app

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"time"

	"toque/internal/netq"
	"toque/internal/roblox"
	"toque/internal/winx"
)

type ServerProbe struct {
	JobID   string  `json:"jobId"`
	IP      string  `json:"ip"`
	RTT     float64 `json:"rttMs"`
	Jitter  float64 `json:"jitterMs"`
	Loss    float64 `json:"lossPct"`
	Playing int     `json:"playing"`
	Max     int     `json:"max"`
	Current bool    `json:"current"`
	Cached  bool    `json:"cached"`
	Error   string  `json:"error,omitempty"`
}

type ServerSearch struct {
	Probes []ServerProbe `json:"probes"`
	Best   *ServerProbe  `json:"best"`
	Joined bool          `json:"joined"`
}

func measureServer(ctx context.Context, ip string, dur time.Duration) (netq.Summary, error) {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return netq.Summary{}, err
	}
	dir := filepath.Join(winx.DataDir("finder"), time.Now().Format("150405.000"))
	defer os.RemoveAll(dir)
	pcap, err := netq.Record(ctx, ip, dur, dir)
	if err != nil {
		return netq.Summary{}, err
	}
	pk, err := netq.ReadPcapng(pcap, addr)
	if err != nil {
		return netq.Summary{}, err
	}
	sum := netq.Analyze(pk)
	if !sum.Confident {
		return sum, errors.New("muy poco tráfico para medir")
	}
	rememberDatacenter(ip, sum.RTT)
	return sum, nil
}

const nearMarginMs = 8

func serverRTT(ctx context.Context, ip string) (ServerProbe, error) {
	p := ServerProbe{IP: ip}
	if rtt, ok := knownDatacenter(ip); ok {
		p.RTT, p.Cached = rtt, true
		return p, nil
	}
	time.Sleep(4 * time.Second)
	sum, err := measureServer(ctx, ip, 8*time.Second)
	p.RTT, p.Jitter, p.Loss = sum.RTT, sum.Jitter, sum.LossDownPct
	return p, err
}

type PlayResult struct {
	First  ServerProbe `json:"first"`
	Final  ServerProbe `json:"final"`
	Hops   int         `json:"hops"`
	BestMs float64     `json:"bestMs"`
}

func (s *Service) Play() (PlayResult, error) {
	ctx, done, err := s.begin("play")
	if err != nil {
		return PlayResult{}, err
	}
	defer done()
	inst, err := roblox.Find()
	if err != nil {
		return PlayResult{}, err
	}
	s.progress("play", "join", 0.08, "Entrando en Realistic Street Soccer")
	sess, err := roblox.Join(ctx, inst, roblox.RealisticStreetSoccer, "", 90*time.Second)
	if err != nil {
		return PlayResult{}, fmt.Errorf("no pude entrar al juego: %w", err)
	}
	s.progress("play", "measure", 0.3, "Comprobando a qué distancia está el servidor")
	first, err := serverRTT(ctx, sess.ServerIP)
	first.JobID, first.Current = sess.JobID, true
	res := PlayResult{First: first, Final: first}
	best, known := bestDatacenter()
	res.BestMs = best
	if err != nil || !known || first.RTT <= best+nearMarginMs {
		return res, nil
	}
	candidates := []roblox.PublicServer{{}}
	if servers, err := roblox.PublicServers(ctx, roblox.RealisticStreetSoccer, 2); err == nil {
		candidates = append(candidates, roblox.SampleCandidates(servers, map[string]bool{sess.JobID: true}, 2, time.Now().UnixNano())...)
	}
	visited := map[string]bool{sess.JobID: true}
	current := sess.JobID
	for i, c := range candidates {
		if ctx.Err() != nil {
			break
		}
		s.progress("play", "hop", 0.4+0.18*float64(i), fmt.Sprintf("El servidor va a %.0f ms y tienes uno a %.0f ms. Probando otro (%d de %d)", first.RTT, best, i+1, len(candidates)))
		ns, err := roblox.Join(ctx, inst, roblox.RealisticStreetSoccer, c.ID, 35*time.Second)
		if err != nil || (c.ID != "" && ns.JobID != c.ID) {
			current = ""
			continue
		}
		current = ns.JobID
		res.Hops++
		if visited[ns.JobID] {
			continue
		}
		visited[ns.JobID] = true
		p, err := serverRTT(ctx, ns.ServerIP)
		if err != nil {
			continue
		}
		p.JobID, p.Playing, p.Max = ns.JobID, c.Playing, c.MaxPlayers
		if p.RTT < res.Final.RTT {
			res.Final = p
		}
		if p.RTT <= best+nearMarginMs {
			return res, nil
		}
	}
	if current != res.Final.JobID {
		s.progress("play", "join", 0.95, fmt.Sprintf("Volviendo al servidor más cercano que encontré, a %.0f ms", res.Final.RTT))
		roblox.Join(ctx, inst, roblox.RealisticStreetSoccer, res.Final.JobID, 35*time.Second)
	}
	return res, nil
}

func (s *Service) QuickPing() (ServerProbe, error) {
	ctx, done, err := s.begin("ping")
	if err != nil {
		return ServerProbe{}, err
	}
	defer done()
	inst, err := roblox.Find()
	if err != nil {
		return ServerProbe{}, err
	}
	sess, err := currentSession(inst)
	if err != nil {
		return ServerProbe{}, errors.New("entra en una partida para medir tu ping")
	}
	s.progress("ping", "measure", 0.3, "Escuchando el tráfico del servidor durante 8 segundos")
	p := ServerProbe{JobID: sess.JobID, IP: sess.ServerIP, Current: true}
	sum, err := measureServer(ctx, sess.ServerIP, 8*time.Second)
	if err != nil {
		return p, err
	}
	p.RTT, p.Jitter, p.Loss = sum.RTT, sum.Jitter, sum.LossDownPct
	return p, nil
}

func (s *Service) FindServer(tries int) (ServerSearch, error) {
	ctx, done, err := s.begin("finder")
	if err != nil {
		return ServerSearch{}, err
	}
	defer done()
	if tries <= 0 || tries > 6 {
		tries = 3
	}
	inst, err := roblox.Find()
	if err != nil {
		return ServerSearch{}, err
	}
	var res ServerSearch
	exclude := map[string]bool{}
	steps := float64(tries + 3)
	step := 0.0
	tick := func(msg string) {
		step++
		s.progress("finder", "probe", step/steps, msg)
	}
	if sess, err := currentSession(inst); err == nil {
		tick("Midiendo el ping del servidor actual")
		exclude[sess.JobID] = true
		p := ServerProbe{JobID: sess.JobID, IP: sess.ServerIP, Current: true}
		if sum, err := measureServer(ctx, sess.ServerIP, 8*time.Second); err == nil {
			p.RTT, p.Jitter, p.Loss = sum.RTT, sum.Jitter, sum.LossDownPct
		} else {
			p.Error = err.Error()
		}
		res.Probes = append(res.Probes, p)
	}
	tick("Consultando la lista pública de servidores")
	servers, err := roblox.PublicServers(ctx, roblox.RealisticStreetSoccer, 2)
	if err != nil && len(servers) == 0 {
		return res, fmt.Errorf("no pude leer la lista de servidores: %w", err)
	}
	for _, c := range roblox.SampleCandidates(servers, exclude, tries, time.Now().UnixNano()) {
		if ctx.Err() != nil {
			break
		}
		tick(fmt.Sprintf("Probando un servidor con %d/%d jugadores", c.Playing, c.MaxPlayers))
		p := ServerProbe{JobID: c.ID, Playing: c.Playing, Max: c.MaxPlayers}
		sess, err := roblox.Join(ctx, inst, roblox.RealisticStreetSoccer, c.ID, 35*time.Second)
		if err != nil || sess.JobID != c.ID {
			p.Error = "no se pudo entrar (servidor lleno o cerrado)"
			res.Probes = append(res.Probes, p)
			continue
		}
		p.IP = sess.ServerIP
		if rtt, ok := knownDatacenter(sess.ServerIP); ok {
			p.RTT, p.Cached = rtt, true
			res.Probes = append(res.Probes, p)
			continue
		}
		time.Sleep(4 * time.Second)
		if sum, err := measureServer(ctx, sess.ServerIP, 8*time.Second); err == nil {
			p.RTT, p.Jitter, p.Loss = sum.RTT, sum.Jitter, sum.LossDownPct
		} else {
			p.Error = err.Error()
		}
		res.Probes = append(res.Probes, p)
	}
	var best *ServerProbe
	for i := range res.Probes {
		p := &res.Probes[i]
		if p.Error == "" && p.RTT > 0 && (best == nil || p.RTT < best.RTT-1) {
			best = p
		}
	}
	if best == nil {
		return res, errors.New("no pude medir ningún servidor")
	}
	res.Best = best
	last := res.Probes[len(res.Probes)-1]
	if last.JobID == best.JobID && last.Error == "" {
		res.Joined = true
		return res, nil
	}
	tick("Entrando al servidor con menos ping")
	if sess, err := roblox.Join(ctx, inst, roblox.RealisticStreetSoccer, best.JobID, 35*time.Second); err == nil && sess.JobID == best.JobID {
		res.Joined = true
	}
	return res, nil
}
