package frames

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Frame struct {
	CPUStartMs           float64
	BetweenPresents      float64
	BetweenDisplayChange float64
	UntilDisplayed       float64
	CPUBusy              float64
	CPUWait              float64
	GPULatency           float64
	GPUBusy              float64
	GPUWait              float64
	ClickToPhoton        float64
	InputToPhoton        float64
	PresentMode          string
	AllowsTearing        bool
}

func Capture(ctx context.Context, presentMon string, pid uint32, dur time.Duration, outCSV string) error {
	os.Remove(outCSV)
	secs := int(math.Ceil(dur.Seconds()))
	cmd := exec.CommandContext(ctx, presentMon,
		"--process_id", strconv.FormatUint(uint64(pid), 10),
		"--output_file", outCSV,
		"--timed", strconv.Itoa(secs),
		"--terminate_after_timed",
		"--no_console_stats",
		"--stop_existing_session",
		"--session_name", "ToqueFrames",
		"--qpc_time_ms",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("presentmon: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if _, err := os.Stat(outCSV); err != nil {
		return errors.New("presentmon produced no csv (process not presenting?)")
	}
	return nil
}

func ParseCSV(path string) ([]Frame, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	idx := map[string]int{}
	for i, h := range header {
		idx[strings.TrimPrefix(strings.TrimSpace(h), string(rune(0xFEFF)))] = i
	}
	num := func(rec []string, name string) float64 {
		i, ok := idx[name]
		if !ok || i >= len(rec) {
			return math.NaN()
		}
		v, err := strconv.ParseFloat(rec[i], 64)
		if err != nil {
			return math.NaN()
		}
		return v
	}
	str := func(rec []string, name string) string {
		if i, ok := idx[name]; ok && i < len(rec) {
			return rec[i]
		}
		return ""
	}
	var out []Frame
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return out, err
		}
		out = append(out, Frame{
			CPUStartMs:           num(rec, "CPUStartQPCTimeInMs"),
			BetweenPresents:      num(rec, "MsBetweenPresents"),
			BetweenDisplayChange: num(rec, "MsBetweenDisplayChange"),
			UntilDisplayed:       num(rec, "MsUntilDisplayed"),
			CPUBusy:              num(rec, "MsCPUBusy"),
			CPUWait:              num(rec, "MsCPUWait"),
			GPULatency:           num(rec, "MsGPULatency"),
			GPUBusy:              num(rec, "MsGPUBusy"),
			GPUWait:              num(rec, "MsGPUWait"),
			ClickToPhoton:        num(rec, "MsClickToPhotonLatency"),
			InputToPhoton:        num(rec, "MsAllInputToPhotonLatency"),
			PresentMode:          str(rec, "PresentMode"),
			AllowsTearing:        str(rec, "AllowsTearing") == "1",
		})
	}
	return out, nil
}
