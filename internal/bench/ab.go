package bench

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"toque/internal/i18n"
	"toque/internal/numfmt"
	"toque/internal/roblox"
	"toque/internal/stats"
	"toque/internal/tweaks"
	"toque/internal/winx"
)

type ABConfig struct {
	Tweak   string        `json:"tweak"`
	Cycles  int           `json:"cycles"`
	Window  time.Duration `json:"window"`
	Warmup  time.Duration `json:"warmup"`
	Settle  time.Duration `json:"settle"`
	PlaceID int64         `json:"placeId"`
	Stress  string        `json:"stress"`
	Primary []string      `json:"primary"`
}

type Sample struct {
	Arm     string             `json:"arm"`
	Valid   bool               `json:"valid"`
	Reason  string             `json:"reason,omitempty"`
	Metrics map[string]float64 `json:"metrics"`
	Window  Window             `json:"window"`
}

type ABResult struct {
	Config   ABConfig                    `json:"config"`
	Kind     tweaks.Kind                 `json:"kind"`
	Started  time.Time                   `json:"started"`
	Finished time.Time                   `json:"finished"`
	Samples  []Sample                    `json:"samples"`
	Compare  map[string]stats.Comparison `json:"compare"`
	Verdict  string                      `json:"verdict"`
	Summary  string                      `json:"summary"`
	Dir      string                      `json:"dir"`
}

var Better = map[string]int{
	"fps": 1, "low1": 1, "low01": 1, "stutterMin": -1, "latMs": -1, "ftStdev": -1,
	"cpuBusy": -1, "gpuBusy": -1, "rtt": -1, "jitter": -1, "lossDown": -1, "maxGap": -1,
	"hardFaults": -1, "ramRoblox": -1, "cpuRoblox": -1,
}

func MetricsOf(w Window) map[string]float64 {
	m := map[string]float64{
		"fps": w.Frames.AvgFPS, "low1": w.Frames.Low1FPS, "low01": w.Frames.Low01FPS,
		"stutterMin": w.Frames.StutterPerMin, "latMs": w.Frames.RenderLatency, "ftStdev": w.Frames.FrametimeStdev,
		"cpuBusy": w.Frames.CPUBusy, "gpuBusy": w.Frames.GPUBusy,
		"hardFaults": w.Sys.HardFaults, "ramRoblox": w.Sys.ProcMB, "cpuRoblox": w.Sys.CPUProc,
	}
	if w.Net.Confident {
		m["rtt"] = w.Net.RTT
		m["jitter"] = w.Net.Jitter
		m["lossDown"] = w.Net.LossDownPct
		m["maxGap"] = w.Net.MaxGapMs
	}
	return m
}

func armSequence(cycles int) []string {
	var seq []string
	for i := 0; i < cycles; i++ {
		if i%2 == 0 {
			seq = append(seq, "A", "B")
		} else {
			seq = append(seq, "B", "A")
		}
	}
	return seq
}

func setArm(ctx tweaks.Context, t tweaks.Tweak, arm string) error {
	if arm == "B" {
		return tweaks.Apply(ctx, t)
	}
	return tweaks.Revert(ctx, t)
}

func RunAB(ctx context.Context, env Env, cfg ABConfig, logf func(string, ...any)) (res ABResult, err error) {
	t, err := tweaks.Get(cfg.Tweak)
	if err != nil {
		return ABResult{}, err
	}
	entries, err := tweaks.Entries()
	if err != nil {
		return ABResult{}, err
	}
	_, wasApplied := entries[t.ID()]
	tctx := tweaks.Context{Install: env.Install}
	if !wasApplied && t.State(tctx).Applied {
		return ABResult{}, errors.New(i18n.T("este cambio ya estaba así antes de Toque, así que no hay un estado sin él con el que comparar"))
	}
	res = ABResult{Config: cfg, Kind: t.Kind(), Started: time.Now()}
	safe := strings.NewReplacer(":", "_", "=", "-", ",", "+", "/", "_", "*", "").Replace(cfg.Tweak)
	res.Dir = filepath.Join(winx.DataDir("experiments"), time.Now().Format("20060102-150405")+"-"+safe)
	os.MkdirAll(res.Dir, 0o755)
	stop := StartStress(cfg.Stress)
	defer stop()
	seq := armSequence(cfg.Cycles)
	var sess roblox.Session
	writeMarker(t.ID(), wasApplied)
	defer func() {
		tctx.PID = sess.PID
		if t.Kind() != tweaks.Live {
			roblox.CloseAll(10 * time.Second)
			tctx.PID = 0
		}
		arm := "A"
		if wasApplied {
			arm = "B"
		}
		if rerr := setArm(tctx, t, arm); rerr != nil {
			if err == nil {
				err = fmt.Errorf(i18n.T("no pude dejar el cambio como estaba y se deshará al volver a abrir Toque (%w)"), rerr)
			}
			return
		}
		clearMarker()
	}()
	joinFresh := func() error {
		s, err := roblox.Join(ctx, env.Install, cfg.PlaceID, sess.JobID, 90*time.Second)
		if err != nil && sess.JobID != "" {
			s, err = roblox.Join(ctx, env.Install, cfg.PlaceID, "", 90*time.Second)
		}
		if err != nil {
			return err
		}
		sess = s
		time.Sleep(cfg.Warmup)
		return nil
	}
	if t.Kind() == tweaks.Live {
		tctx.PID = 0
		if err := setArm(tctx, t, "A"); err != nil {
			return res, err
		}
		if err := joinFresh(); err != nil {
			return res, err
		}
	}
	for i, arm := range seq {
		if ctx.Err() != nil {
			break
		}
		if t.Kind() == tweaks.Live {
			tctx.PID = sess.PID
			if err := setArm(tctx, t, arm); err != nil {
				return res, fmt.Errorf("set arm %s: %w", arm, err)
			}
			time.Sleep(cfg.Settle)
		} else {
			roblox.CloseAll(10 * time.Second)
			tctx.PID = 0
			if err := setArm(tctx, t, arm); err != nil {
				return res, fmt.Errorf("set arm %s: %w", arm, err)
			}
			if err := joinFresh(); err != nil {
				res.Samples = append(res.Samples, Sample{Arm: arm, Reason: "join failed: " + err.Error()})
				continue
			}
		}
		hwnd := winx.MainWindow(sess.PID)
		winx.Focus(hwnd)
		w := Measure(ctx, env, sess, cfg.Window, filepath.Join(res.Dir, fmt.Sprintf("%02d-%s", i+1, arm)), arm)
		smp := Sample{Arm: arm, Valid: true, Metrics: MetricsOf(w), Window: w}
		switch {
		case !w.Foreground:
			smp.Valid, smp.Reason = false, "roblox lost focus"
		case winx.IdleTime() < cfg.Window-2*time.Second:
			smp.Valid, smp.Reason = false, "user input during window"
		case w.Frames.Frames == 0:
			smp.Valid, smp.Reason = false, "no frames captured"
		}
		res.Samples = append(res.Samples, smp)
		if logf != nil {
			logf("%2d/%d %s valid=%v fps=%.1f low1=%.1f low01=%.1f stut=%.1f lat=%.2f rtt=%.1f %s", i+1, len(seq), arm, smp.Valid,
				w.Frames.AvgFPS, w.Frames.Low1FPS, w.Frames.Low01FPS, w.Frames.StutterPerMin, w.Frames.RenderLatency, w.Net.RTT, smp.Reason)
		}
	}
	roblox.CloseAll(10 * time.Second)
	sess.PID = 0
	res.Finished = time.Now()
	res.Compare, res.Verdict, res.Summary = Evaluate(res.Samples, cfg.Primary)
	b, _ := json.MarshalIndent(res, "", "  ")
	os.WriteFile(filepath.Join(res.Dir, "result.json"), b, 0o644)
	return res, nil
}

func Evaluate(samples []Sample, primary []string) (map[string]stats.Comparison, string, string) {
	a := map[string][]float64{}
	b := map[string][]float64{}
	for _, s := range samples {
		if !s.Valid {
			continue
		}
		for k, v := range s.Metrics {
			if s.Arm == "A" {
				a[k] = append(a[k], v)
			} else {
				b[k] = append(b[k], v)
			}
		}
	}
	primaryAlpha := 0.05 / float64(max(1, len(primary)))
	cmp := map[string]stats.Comparison{}
	for k := range Better {
		if len(a[k]) == 0 || len(b[k]) == 0 {
			continue
		}
		alpha := 0.01
		if contains(primary, k) {
			alpha = primaryAlpha
		}
		cmp[k] = stats.CompareAlpha(a[k], b[k], 42, alpha)
	}
	improved, worsened := []string{}, []string{}
	for _, k := range primary {
		c, ok := cmp[k]
		if !ok || !c.Conclusive {
			continue
		}
		if c.DeltaPct*float64(Better[k]) > 0 {
			improved = append(improved, describe(k, c.DeltaPct))
		} else {
			worsened = append(worsened, describe(k, c.DeltaPct))
		}
	}
	guardWorse := []string{}
	for k, c := range cmp {
		if !c.Conclusive || contains(primary, k) {
			continue
		}
		if c.DeltaPct*float64(Better[k]) < -2 {
			guardWorse = append(guardWorse, describe(k, c.DeltaPct))
		}
	}
	sort.Strings(guardWorse)
	switch {
	case len(worsened) > 0:
		return cmp, "worse", i18n.Tf("Empeora %s", join(worsened))
	case len(improved) > 0 && len(guardWorse) == 0:
		return cmp, "improves", i18n.Tf("Mejora %s", join(improved))
	case len(improved) > 0:
		return cmp, "tradeoff", i18n.Tf("Mejora %s, pero empeora %s", join(improved), join(guardWorse))
	default:
		labels := make([]string, 0, len(primary))
		for _, k := range primary {
			labels = append(labels, label(k))
		}
		return cmp, "no-effect", i18n.Tf("Sin efecto medible en %s", join(labels))
	}
}

var labels = map[string]string{
	"fps": i18n.Mark("los FPS medios"), "low1": i18n.Mark("el 1 % low"), "low01": i18n.Mark("el 0,1 % low"), "stutterMin": i18n.Mark("los tirones"),
	"latMs": i18n.Mark("la latencia de render"), "ftStdev": i18n.Mark("la variación del frametime"), "cpuBusy": i18n.Mark("la CPU por frame"),
	"gpuBusy": i18n.Mark("la GPU por frame"), "rtt": i18n.Mark("el ping"), "jitter": i18n.Mark("el jitter"), "lossDown": i18n.Mark("la pérdida de paquetes"),
	"maxGap": i18n.Mark("el mayor corte de red"), "hardFaults": i18n.Mark("las lecturas de paginación"), "ramRoblox": i18n.Mark("la RAM de Roblox"),
	"cpuRoblox": i18n.Mark("la CPU de Roblox"),
}

func label(k string) string {
	if l, ok := labels[k]; ok {
		return i18n.T(l)
	}
	return k
}

func describe(k string, delta float64) string {
	return fmt.Sprintf("%s (%s)", label(k), numfmt.Localize(fmt.Sprintf("%+.1f %%", delta)))
}

func join(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	}
	return strings.Join(items[:len(items)-1], ", ") + " " + i18n.T("y") + " " + items[len(items)-1]
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
