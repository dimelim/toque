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
	"time"

	"toque/internal/bench"
	"toque/internal/frames"
	"toque/internal/netq"
	"toque/internal/roblox"
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
		pm = `.dev\bin\PresentMon.exe`
	}
	return bench.Env{PresentMon: pm, Install: inst, Dir: winx.DataDir("runs")}, nil
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
	fs.Parse(args)
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
