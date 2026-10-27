package netq

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func pktmon(args ...string) (string, error) {
	cmd := exec.Command("pktmon", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

type Recorder struct {
	dir  string
	etl  string
	pcap string
}

func Start(serverIP string, dir string) (*Recorder, error) {
	os.MkdirAll(dir, 0o755)
	r := &Recorder{dir: dir, etl: filepath.Join(dir, "net.etl"), pcap: filepath.Join(dir, "net.pcapng")}
	os.Remove(r.etl)
	os.Remove(r.pcap)
	pktmon("stop")
	pktmon("filter", "remove")
	if out, err := pktmon("filter", "add", "ToqueRbx", "-i", serverIP, "-t", "UDP"); err != nil {
		return nil, fmt.Errorf("pktmon filter: %v: %s", err, out)
	}
	if out, err := pktmon("start", "--capture", "--comp", "nics", "--pkt-size", "128", "--file-name", r.etl); err != nil {
		pktmon("filter", "remove")
		return nil, fmt.Errorf("pktmon start: %v: %s", err, out)
	}
	return r, nil
}

func (r *Recorder) Stop() (string, error) {
	defer pktmon("filter", "remove")
	if out, err := pktmon("stop"); err != nil {
		return "", fmt.Errorf("pktmon stop: %v: %s", err, out)
	}
	if out, err := pktmon("etl2pcap", r.etl, "--out", r.pcap); err != nil {
		return "", fmt.Errorf("pktmon etl2pcap: %v: %s", err, out)
	}
	os.Remove(r.etl)
	return r.pcap, nil
}

func Record(ctx context.Context, serverIP string, dur time.Duration, dir string) (string, error) {
	r, err := Start(serverIP, dir)
	if err != nil {
		return "", err
	}
	select {
	case <-ctx.Done():
	case <-time.After(dur):
	}
	return r.Stop()
}
