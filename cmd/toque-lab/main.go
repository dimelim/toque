package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"toque/internal/bench"
	"toque/internal/frames"
	"toque/internal/netq"
	"toque/internal/roblox"
	"toque/internal/tweaks"
	"toque/internal/winx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: toque-lab <status|join|measure|trial|analyze-pcap|analyze-csv> [flags]")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "status":
		err = cmdStatus()
	case "join":
		err = cmdJoin(os.Args[2:])
	case "measure":
		err = cmdMeasure(os.Args[2:])
	case "trial":
		err = cmdTrial(os.Args[2:])
	case "trials":
		err = cmdTrials(os.Args[2:])
	case "tweak":
		err = cmdTweak(os.Args[2:])
	case "ab":
		err = cmdAB(os.Args[2:])
	case "report":
		err = cmdReport(os.Args[2:])
	case "bloat":
		err = cmdBloat(os.Args[2:])
	case "finder":
		err = cmdFinder(os.Args[2:])
	case "load":
		err = cmdLoad(os.Args[2:])
	case "play":
		err = cmdPlay(os.Args[2:])
	case "regions":
		err = cmdRegions(os.Args[2:])
	case "diagnose":
		err = cmdDiagnose(os.Args[2:])
	case "analyze-pcap":
		err = cmdAnalyzePcap(os.Args[2:])
	case "analyze-csv":
		err = cmdAnalyzeCSV(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func printJSON(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func env() (bench.Env, error) {
	inst, err := roblox.Find()
	if err != nil {
		return bench.Env{}, err
	}
	exe, _ := os.Executable()
	pm := filepath.Join(filepath.Dir(exe), "PresentMon.exe")
	if _, err := os.Stat(pm); err != nil {
		pm = filepath.Join(".dev", "bin", "PresentMon.exe")
	}
	return bench.Env{PresentMon: pm, Install: inst, Dir: winx.DataDir("runs"), KeepRaw: true}, nil
}

func currentSession(inst roblox.Install) (roblox.Session, error) {
	pids := roblox.Running()
	if len(pids) == 0 {
		return roblox.Session{}, errors.New("roblox is not running")
	}
	log := roblox.NewestPlayerLog(inst.LogDir(), time.Now().Add(-24*time.Hour))
	if log == "" {
		return roblox.Session{}, errors.New("no player log found")
	}
	s := roblox.Session{PID: pids[0], LogPath: log}
	for _, ev := range roblox.ScanFile(log) {
		switch ev.Kind {
		case roblox.EventJoining:
			s.JobID, s.PlaceID = ev.JobID, ev.PlaceID
			s.ServerIP = ""
		case roblox.EventUdmux:
			s.ServerIP, s.ServerPort = ev.IP, ev.Port
		case roblox.EventLoadTime:
			s.JoinSecs = ev.Seconds
		}
	}
	if s.ServerIP == "" {
		return s, errors.New("roblox is running but not connected to a server")
	}
	return s, nil
}

func cmdStatus() error {
	e, err := env()
	out := map[string]any{"elevated": winx.IsElevated(), "running": roblox.Running()}
	if err != nil {
		out["installError"] = err.Error()
		printJSON(out)
		return nil
	}
	out["install"] = e.Install
	out["presentMon"] = e.PresentMon
	if st, err := roblox.LoadBasicSettings(e.Install.SettingsPath()); err == nil {
		fc, _ := st.Int("FramerateCap")
		q, _ := st.Token("SavedQualityLevel")
		gm, _ := st.Token("GraphicsOptimizationMode")
		fs, _ := st.Bool("Fullscreen")
		ps, _ := st.Bool("PerformanceStatsVisible")
		out["settings"] = map[string]any{"framerateCap": fc, "savedQuality": q, "graphicsMode": gm, "fullscreen": fs, "perfStats": ps}
	}
	if fl, err := roblox.ReadFlags(e.Install.VersionDir); err == nil {
		out["flags"] = fl
	}
	if s, err := currentSession(e.Install); err == nil {
		out["session"] = s
	}
	printJSON(out)
	return nil
}

func cmdJoin(args []string) error {
	fs := flag.NewFlagSet("join", flag.ExitOnError)
	place := fs.Int64("place", roblox.RealisticStreetSoccer, "place id")
	job := fs.String("job", "", "server job id")
	fs.Parse(args)
	e, err := env()
	if err != nil {
		return err
	}
	s, err := roblox.Join(context.Background(), e.Install, *place, *job, 90*time.Second)
	printJSON(map[string]any{"session": s, "timeToJoinMs": s.TimeToJoin().Milliseconds()})
	return err
}

func cmdMeasure(args []string) error {
	fs := flag.NewFlagSet("measure", flag.ExitOnError)
	dur := fs.Duration("dur", 30*time.Second, "duration")
	out := fs.String("out", "", "output dir")
	label := fs.String("label", "measure", "label")
	focus := fs.Bool("focus", true, "bring roblox to foreground first")
	load := fs.String("load", "", "network load during the window: down or up")
	fs.Parse(args)
	if *load != "" {
		stop := bench.StartLoad(*load, 4)
		defer func() { fmt.Fprintf(os.Stderr, "load %s transferred %.0f MB\n", *load, float64(stop())/1e6) }()
		time.Sleep(2 * time.Second)
	}
	e, err := env()
	if err != nil {
		return err
	}
	s, err := currentSession(e.Install)
	if err != nil {
		return err
	}
	if *focus {
		winx.Focus(winx.MainWindow(s.PID))
	}
	dir := *out
	if dir == "" {
		dir = filepath.Join(e.Dir, time.Now().Format("20060102-150405")+"-"+*label)
	}
	printJSON(bench.Measure(context.Background(), e, s, *dur, dir, *label))
	return nil
}

func cmdTrial(args []string) error {
	fs := flag.NewFlagSet("trial", flag.ExitOnError)
	dur := fs.Duration("dur", 30*time.Second, "measure duration")
	warm := fs.Duration("warmup", 15*time.Second, "warmup after join")
	place := fs.Int64("place", roblox.RealisticStreetSoccer, "place id")
	job := fs.String("job", "", "server job id")
	out := fs.String("out", "", "output dir")
	label := fs.String("label", "trial", "label")
	keep := fs.Bool("keep", false, "keep roblox open afterwards")
	fs.Parse(args)
	e, err := env()
	if err != nil {
		return err
	}
	s, err := roblox.Join(context.Background(), e.Install, *place, *job, 90*time.Second)
	if err != nil {
		return err
	}
	time.Sleep(*warm)
	winx.Focus(winx.MainWindow(s.PID))
	dir := *out
	if dir == "" {
		dir = filepath.Join(e.Dir, time.Now().Format("20060102-150405")+"-"+*label)
	}
	w := bench.Measure(context.Background(), e, s, *dur, dir, *label)
	if !*keep {
		roblox.CloseAll(10 * time.Second)
	}
	printJSON(map[string]any{"session": s, "timeToJoinMs": s.TimeToJoin().Milliseconds(), "window": w})
	return nil
}

func cmdTrials(args []string) error {
	fs := flag.NewFlagSet("trials", flag.ExitOnError)
	n := fs.Int("n", 3, "number of trials")
	dur := fs.Duration("dur", 30*time.Second, "measure duration")
	warm := fs.Duration("warmup", 15*time.Second, "warmup after join")
	place := fs.Int64("place", roblox.RealisticStreetSoccer, "place id")
	label := fs.String("label", "series", "label")
	fs.Parse(args)
	e, err := env()
	if err != nil {
		return err
	}
	base := filepath.Join(e.Dir, time.Now().Format("20060102-150405")+"-"+*label)
	var rows []map[string]any
	for i := 1; i <= *n; i++ {
		s, err := roblox.Join(context.Background(), e.Install, *place, "", 90*time.Second)
		if err != nil {
			rows = append(rows, map[string]any{"trial": i, "error": err.Error()})
			roblox.CloseAll(10 * time.Second)
			continue
		}
		time.Sleep(*warm)
		winx.Focus(winx.MainWindow(s.PID))
		w := bench.Measure(context.Background(), e, s, *dur, filepath.Join(base, fmt.Sprint(i)), *label)
		roblox.CloseAll(10 * time.Second)
		row := map[string]any{
			"trial": i, "server": s.ServerIP, "joinMs": s.TimeToJoin().Milliseconds(), "fg": w.Foreground,
			"fps": round(w.Frames.AvgFPS), "low1": round(w.Frames.Low1FPS), "low01": round(w.Frames.Low01FPS),
			"stutterMin": round(w.Frames.StutterPerMin), "latMs": round(w.Frames.RenderLatency),
			"rtt": round(w.Net.RTT), "jit": round(w.Net.Jitter), "loss": round(w.Net.LossDownPct),
			"hardF": round(w.Sys.HardFaults), "hardFMax": round(w.Sys.HardFaultsMax), "availMB": round(w.Sys.AvailMB),
			"disks": w.Sys.Disks, "errors": w.Errors,
		}
		rows = append(rows, row)
		time.Sleep(3 * time.Second)
	}
	b, _ := json.MarshalIndent(rows, "", "  ")
	os.WriteFile(filepath.Join(base, "series.json"), b, 0o644)
	fmt.Println(string(b))
	return nil
}

func cmdTweak(args []string) error {
	if len(args) == 0 {
		return errors.New("tweak <list|state|apply|revert|revert-all> [id]")
	}
	e, err := env()
	if err != nil {
		return err
	}
	tctx := tweaks.Context{Install: e.Install}
	if pids := roblox.Running(); len(pids) > 0 {
		tctx.PID = pids[0]
	}
	entries, err := tweaks.Entries()
	if err != nil {
		return err
	}
	switch args[0] {
	case "list":
		printJSON(map[string]any{"registered": tweaks.IDs(), "journal": entries})
		return nil
	case "revert-all":
		return tweaks.RevertAll(tctx)
	}
	if len(args) < 2 {
		return errors.New("missing tweak id")
	}
	t, err := tweaks.Get(args[1])
	if err != nil {
		return err
	}
	switch args[0] {
	case "state":
		_, managed := entries[t.ID()]
		printJSON(map[string]any{"id": t.ID(), "kind": t.Kind(), "state": t.State(tctx), "inJournal": managed})
	case "apply":
		err = tweaks.Apply(tctx, t)
		printJSON(t.State(tctx))
	case "revert":
		err = tweaks.Revert(tctx, t)
		printJSON(t.State(tctx))
	default:
		err = errors.New("unknown tweak action")
	}
	return err
}

func cmdAB(args []string) error {
	fs := flag.NewFlagSet("ab", flag.ExitOnError)
	cfg := bench.ABConfig{}
	fs.StringVar(&cfg.Tweak, "tweak", "", "tweak id")
	fs.IntVar(&cfg.Cycles, "cycles", 4, "ABBA cycles (each cycle = 1 A + 1 B)")
	fs.DurationVar(&cfg.Window, "window", 20*time.Second, "measurement window")
	fs.DurationVar(&cfg.Warmup, "warmup", 15*time.Second, "warmup after join")
	fs.DurationVar(&cfg.Settle, "settle", 3*time.Second, "settle after live toggle")
	fs.Int64Var(&cfg.PlaceID, "place", roblox.RealisticStreetSoccer, "place id")
	fs.StringVar(&cfg.Stress, "stress", "", "stress spec e.g. cpu:6,ram:1000,up")
	primary := fs.String("primary", "low1,low01,stutterMin,latMs", "primary metrics")
	fs.Parse(args)
	cfg.Primary = strings.Split(*primary, ",")
	e, err := env()
	if err != nil {
		return err
	}
	res, err := bench.RunAB(context.Background(), e, cfg, func(f string, a ...any) { fmt.Printf(f+"\n", a...) })
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(res.Compare))
	for k := range res.Compare {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		c := res.Compare[k]
		fmt.Printf("%-11s A=%9.2f B=%9.2f d=%+6.1f%% CI[%+6.1f,%+6.1f] p=%.3f %v\n", k, c.MedianA, c.MedianB, c.DeltaPct, c.CILowPct, c.CIHighPct, c.PValue, c.Conclusive)
	}
	fmt.Println("VERDICT:", res.Verdict, "-", res.Summary)
	fmt.Println("dir:", res.Dir)
	return nil
}

func cmdBloat(args []string) error {
	fs := flag.NewFlagSet("bloat", flag.ExitOnError)
	target := fs.String("target", "", "ip to ping (default: first public hop toward roblox)")
	phase := fs.Duration("phase", 15*time.Second, "duration of each phase")
	fs.Parse(args)
	dst, err := netip.ParseAddr(*target)
	if err != nil {
		hops := netq.Trace(netip.MustParseAddr("128.116.127.33"), 8, 2)
		for _, h := range hops {
			if !h.Private {
				dst = netip.MustParseAddr(h.Addr)
				break
			}
		}
		if !dst.IsValid() && len(hops) > 0 {
			dst = netip.MustParseAddr(hops[0].Addr)
		}
	}
	if !dst.IsValid() {
		return errors.New("no target to ping")
	}
	res := netq.Bufferbloat(context.Background(), dst, *phase, func(kind string) func() int64 { return bench.StartLoad(kind, 4) })
	printJSON(res)
	return nil
}

func round(v float64) float64 {
	return float64(int64(v*10+0.5)) / 10
}

func cmdAnalyzePcap(args []string) error {
	if len(args) < 2 {
		return errors.New("analyze-pcap <file> <serverIP>")
	}
	addr, err := netip.ParseAddr(args[1])
	if err != nil {
		return err
	}
	pk, err := netq.ReadPcapng(args[0], addr)
	if err != nil {
		return err
	}
	printJSON(netq.Analyze(pk))
	return nil
}

func cmdAnalyzeCSV(args []string) error {
	if len(args) < 1 {
		return errors.New("analyze-csv <file>")
	}
	fr, err := frames.ParseCSV(args[0])
	if err != nil {
		return err
	}
	printJSON(frames.Summarize(fr))
	return nil
}
