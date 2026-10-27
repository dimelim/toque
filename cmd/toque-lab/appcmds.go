package main

import (
	"flag"
	"fmt"
	"time"

	"toque/internal/app"
	"toque/internal/bench"
)

func newService() (*app.Service, error) {
	e, err := env()
	if err != nil {
		return nil, err
	}
	return app.New(e.PresentMon, func(name string, data any) {
		if p, ok := data.(app.Progress); ok && p.Message != "" {
			fmt.Printf("[%s %.0f%%] %s\n", p.Task, p.Pct*100, p.Message)
		}
	}), nil
}

func cmdFinder(args []string) error {
	fs := flag.NewFlagSet("finder", flag.ExitOnError)
	tries := fs.Int("tries", 3, "servers to probe")
	fs.Parse(args)
	svc, err := newService()
	if err != nil {
		return err
	}
	res, err := svc.FindServer(*tries)
	printJSON(res)
	return err
}

func cmdPlay([]string) error {
	svc, err := newService()
	if err != nil {
		return err
	}
	res, err := svc.Play()
	printJSON(res)
	return err
}

func cmdLoad(args []string) error {
	fs := flag.NewFlagSet("load", flag.ExitOnError)
	kind := fs.String("kind", "up", "up or down")
	secs := fs.Int("secs", 8, "seconds")
	fs.Parse(args)
	stop := bench.StartLoad(*kind, 4)
	time.Sleep(time.Duration(*secs) * time.Second)
	bytes := stop()
	fmt.Printf("%s %.1f Mbps\n", *kind, float64(bytes)*8/1e6/float64(*secs))
	return nil
}

func cmdDiagnose(args []string) error {
	fs := flag.NewFlagSet("diagnose", flag.ExitOnError)
	game := fs.Bool("game", false, "measure inside the game")
	fs.Parse(args)
	svc, err := newService()
	if err != nil {
		return err
	}
	d, err := svc.Diagnose(*game)
	printJSON(d.Findings)
	printJSON(d.Path)
	return err
}
