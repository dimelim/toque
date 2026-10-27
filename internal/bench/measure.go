package bench

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sync"
	"time"

	"toque/internal/frames"
	"toque/internal/netq"
	"toque/internal/roblox"
	"toque/internal/sysmon"
	"toque/internal/winx"
)

type Env struct {
	PresentMon string
	Install    roblox.Install
	Dir        string
	KeepRaw    bool
}

type Window struct {
	Label      string         `json:"label"`
	Start      time.Time      `json:"start"`
	Seconds    float64        `json:"seconds"`
	JobID      string         `json:"jobId"`
	ServerIP   string         `json:"serverIp"`
	Foreground bool           `json:"foreground"`
	Frames     frames.Summary `json:"frames"`
	Net        netq.Summary   `json:"net"`
	Sys        sysmon.Summary `json:"sys"`
	Errors     []string       `json:"errors,omitempty"`
}

func Measure(ctx context.Context, env Env, sess roblox.Session, dur time.Duration, dir string, label string) Window {
	os.MkdirAll(dir, 0o755)
	w := Window{Label: label, Start: time.Now(), Seconds: dur.Seconds(), JobID: sess.JobID, ServerIP: sess.ServerIP}
	hwnd := winx.MainWindow(sess.PID)
	w.Foreground = winx.IsForeground(hwnd)
	var mu sync.Mutex
	fail := func(err error) {
		mu.Lock()
		w.Errors = append(w.Errors, err.Error())
		mu.Unlock()
	}
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		csv := filepath.Join(dir, "frames.csv")
		if err := frames.Capture(ctx, env.PresentMon, sess.PID, dur, csv); err != nil {
			fail(err)
			return
		}
		fr, err := frames.ParseCSV(csv)
		if !env.KeepRaw {
			os.Remove(csv)
		}
		if err != nil {
			fail(err)
			return
		}
		w.Frames = frames.Summarize(fr)
	}()
	go func() {
		defer wg.Done()
		addr, err := netip.ParseAddr(sess.ServerIP)
		if err != nil {
			fail(fmt.Errorf("server ip: %w", err))
			return
		}
		pcap, err := netq.Record(ctx, sess.ServerIP, dur, dir)
		if err != nil {
			fail(err)
			return
		}
		pk, err := netq.ReadPcapng(pcap, addr)
		if !env.KeepRaw {
			os.Remove(pcap)
		}
		if err != nil {
			fail(err)
			return
		}
		w.Net = netq.Analyze(pk)
	}()
	go func() {
		defer wg.Done()
		sctx, cancel := context.WithTimeout(ctx, dur)
		defer cancel()
		samples, err := sysmon.Run(sctx, sess.PID, time.Second)
		if err != nil {
			fail(err)
			return
		}
		w.Sys = sysmon.Summarize(samples)
	}()
	wg.Wait()
	w.Foreground = w.Foreground && winx.IsForeground(hwnd)
	b, _ := json.MarshalIndent(w, "", "  ")
	os.WriteFile(filepath.Join(dir, "window.json"), b, 0o644)
	return w
}
