package main

import (
	"flag"
	"fmt"

	"toque/internal/app"
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
