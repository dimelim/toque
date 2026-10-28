package app

import (
	"time"

	"toque/internal/eco"
	"toque/internal/sysmon"
	"toque/internal/tweaks"
	"toque/internal/winx"
)

const backgroundID = "session:background-eco"

type backgroundTweak struct{}

func (backgroundTweak) ID() string        { return backgroundID }
func (backgroundTweak) Kind() tweaks.Kind { return tweaks.Live }

func (backgroundTweak) State(tweaks.Context) tweaks.State {
	return tweaks.State{Applicable: winx.IsElevated(), Applied: len(throttler().Throttled()) > 0}
}

func (backgroundTweak) Snapshot(tweaks.Context) (tweaks.Undo, error) {
	return tweaks.Undo{}, nil
}

func (backgroundTweak) Apply(tweaks.Context) error {
	var sampler sysmon.CPUSampler
	sampler.Sample()
	time.Sleep(3 * time.Second)
	procs := sampler.Sample()
	audible, err := winx.AudiblePIDs()
	if err != nil {
		return err
	}
	session, _ := winx.SessionOf(ownPID())
	throttler().Apply(eco.Pick(procs, session, audible, nil))
	return nil
}

func (backgroundTweak) Revert(tweaks.Context, tweaks.Undo) error {
	return throttler().RestoreAll()
}

func init() {
	tweaks.Register(backgroundTweak{})
}

func latestExperiments(ids ...string) map[string]*ExperimentRecord {
	out := map[string]*ExperimentRecord{}
	for _, id := range ids {
		out[id] = nil
	}
	for _, r := range LoadExperiments() {
		if best, ok := out[r.Tweak]; ok && (best == nil || r.At.After(best.At)) {
			r := r
			out[r.Tweak] = &r
		}
	}
	return out
}
