package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"toque/internal/bench"
	"toque/internal/catalog"
	"toque/internal/roblox"
	"toque/internal/tweaks"
)

type Optimization struct {
	catalog.Entry
	Kind       string            `json:"kind"`
	Applicable bool              `json:"applicable"`
	Applied    bool              `json:"applied"`
	Managed    bool              `json:"managed"`
	Detail     string            `json:"detail"`
	Mine       *ExperimentRecord `json:"mine"`
}

func (s *Service) context() tweaks.Context {
	ctx := tweaks.Context{}
	if inst, err := roblox.Find(); err == nil {
		ctx.Install = inst
	}
	if pids := roblox.Running(); len(pids) > 0 {
		ctx.PID = pids[0]
	}
	return ctx
}

func (s *Service) describe(e catalog.Entry, managed map[string]tweaks.Undo, exps []ExperimentRecord) Optimization {
	o := Optimization{Entry: e}
	t, err := tweaks.Get(e.ID)
	if err != nil {
		o.Detail = err.Error()
		return o
	}
	st := t.State(s.context())
	o.Kind = string(t.Kind())
	o.Applicable, o.Applied, o.Detail = st.Applicable, st.Applied, st.Detail
	_, o.Managed = managed[e.ID]
	if strings.HasPrefix(e.ID, "flag:") {
		o.Managed = tweaks.FlagsManaged(managed, e.ID)
	}
	for i := range exps {
		if exps[i].Tweak == e.ID {
			o.Mine = &exps[i]
			break
		}
	}
	return o
}

func (s *Service) Optimizations() []Optimization {
	managed, _ := tweaks.Entries()
	exps := LoadExperiments()
	out := make([]Optimization, 0, len(catalog.Catalog))
	for _, e := range catalog.Catalog {
		if e.Category == catalog.CategorySession {
			continue
		}
		out = append(out, s.describe(e, managed, exps))
	}
	return out
}

func (s *Service) find(id string) (catalog.Entry, tweaks.Tweak, error) {
	e, ok := catalog.Find(id)
	if !ok {
		return catalog.Entry{}, nil, fmt.Errorf("optimización desconocida %q", id)
	}
	t, err := tweaks.Get(id)
	return e, t, err
}

func (s *Service) Apply(id string) (Optimization, error) {
	if err := s.idle(); err != nil {
		return Optimization{}, err
	}
	e, t, err := s.find(id)
	if err != nil {
		return Optimization{}, err
	}
	if e.Verdict == catalog.VerdictRejected {
		return Optimization{}, errors.New("esta optimización se descartó porque no mejoró nada en las pruebas")
	}
	if t.Kind() == tweaks.Relaunch && len(roblox.Running()) > 0 {
		return Optimization{}, errors.New("cierra Roblox para aplicar este cambio")
	}
	if err := tweaks.Apply(s.context(), t); err != nil {
		return Optimization{}, err
	}
	managed, _ := tweaks.Entries()
	return s.describe(e, managed, LoadExperiments()), nil
}

func (s *Service) Revert(id string) (Optimization, error) {
	if err := s.idle(); err != nil {
		return Optimization{}, err
	}
	e, t, err := s.find(id)
	if err != nil {
		return Optimization{}, err
	}
	if t.Kind() == tweaks.Relaunch && len(roblox.Running()) > 0 {
		return Optimization{}, errors.New("cierra Roblox para revertir este cambio")
	}
	if err := tweaks.Revert(s.context(), t); err != nil {
		return Optimization{}, err
	}
	managed, _ := tweaks.Entries()
	return s.describe(e, managed, LoadExperiments()), nil
}

func (s *Service) ApplyProven() ([]string, error) {
	var applied []string
	var errs []string
	for _, o := range s.Optimizations() {
		if o.Verdict != catalog.VerdictProven || !o.Applicable || o.Applied {
			continue
		}
		if _, err := s.Apply(o.ID); err != nil {
			errs = append(errs, o.Title+": "+err.Error())
			continue
		}
		applied = append(applied, o.ID)
	}
	if len(errs) > 0 {
		return applied, errors.New(strings.Join(errs, "; "))
	}
	return applied, nil
}

func (s *Service) RevertAll() error {
	if err := s.idle(); err != nil {
		return err
	}
	if len(roblox.Running()) > 0 {
		return errors.New("cierra Roblox para revertir todo")
	}
	return errors.Join(throttler().RestoreAll(), restoreSystem(), tweaks.RevertAll(s.context()))
}

func (s *Service) Verify(id string) (ExperimentRecord, error) {
	if _, _, err := s.find(id); err != nil {
		return ExperimentRecord{}, err
	}
	ctx, done, err := s.begin("verify")
	if err != nil {
		return ExperimentRecord{}, err
	}
	defer done()
	env, err := s.env()
	if err != nil {
		return ExperimentRecord{}, err
	}
	cycles := 6
	cfg := bench.ABConfig{
		Tweak: id, Cycles: cycles, Window: 25 * time.Second, Warmup: 15 * time.Second, Settle: 3 * time.Second,
		PlaceID: roblox.RealisticStreetSoccer, Primary: []string{"low1", "low01", "stutterMin", "latMs"},
	}
	total := float64(cycles * 2)
	n := 0.0
	res, err := bench.RunAB(ctx, env, cfg, func(f string, a ...any) {
		n++
		s.progress("verify", "window", n/total, fmt.Sprintf(f, a...))
	})
	if err != nil {
		return ExperimentRecord{}, err
	}
	return toRecord(res), nil
}

func (s *Service) Experiments() []ExperimentRecord {
	return LoadExperiments()
}

func (s *Service) ExperimentGroups() []ExperimentGroup {
	return LoadExperimentGroups()
}
