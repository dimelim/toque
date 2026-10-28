package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"toque/internal/bench"
	"toque/internal/i18n"
	"toque/internal/roblox"
	"toque/internal/tweaks"
	"toque/internal/winx"
)

func (s *Service) Benchmark(label string) (RunRecord, error) {
	ctx, done, err := s.begin("benchmark")
	if err != nil {
		return RunRecord{}, err
	}
	defer done()
	env, err := s.env()
	if err != nil {
		return RunRecord{}, err
	}
	if label == "" {
		label = "Prueba " + time.Now().Format("02/01 15:04")
	}
	s.progress("benchmark", "join", 0.05, i18n.T("Abriendo Realistic Street Soccer"))
	sess, err := roblox.Join(ctx, env.Install, roblox.RealisticStreetSoccer, "", 90*time.Second)
	if err != nil {
		return RunRecord{}, fmt.Errorf(i18n.T("no pude entrar al juego (%w)"), err)
	}
	defer roblox.CloseAll(10 * time.Second)
	s.progress("benchmark", "warmup", 0.25, i18n.T("Dejando que el juego cargue"))
	select {
	case <-ctx.Done():
		return RunRecord{}, ctx.Err()
	case <-time.After(15 * time.Second):
	}
	winx.Focus(winx.MainWindow(sess.PID))
	s.progress("benchmark", "measure", 0.4, i18n.T("Midiendo frames, red y sistema durante 30 s"))
	id := time.Now().Format("20060102-150405")
	w := bench.Measure(ctx, env, sess, 30*time.Second, filepath.Join(env.Dir, id), label)
	if w.Net.Confident {
		rememberDatacenter(sess.ServerIP, w.Net.RTT)
	}
	entries, _ := tweaks.Entries()
	var applied []string
	for k := range entries {
		applied = append(applied, k)
	}
	rec := RunRecord{ID: id, Label: label, At: time.Now(), ServerIP: sess.ServerIP, JobID: sess.JobID, JoinMs: sess.TimeToJoin().Milliseconds(), Applied: applied, Window: w}
	if err := SaveRun(rec); err != nil {
		return rec, err
	}
	return localRun(rec), nil
}

func localRun(r RunRecord) RunRecord {
	r.Label = localRunLabel(r.Label)
	r.Window.Label = localRunLabel(r.Window.Label)
	return r
}

func localRunLabel(label string) string {
	if date, ok := strings.CutPrefix(label, "Prueba "); ok {
		return i18n.Tf("Prueba %s", date)
	}
	return label
}

func (s *Service) Runs() []RunRecord {
	runs := LoadRuns()
	for i := range runs {
		runs[i] = localRun(runs[i])
	}
	return runs
}

func (s *Service) RemoveRun(id string) error {
	return DeleteRun(id)
}
