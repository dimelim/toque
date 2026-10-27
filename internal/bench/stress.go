package bench

import (
	"context"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"toque/internal/sysmon"
)

const ramFloorMB = 1536

func StartStress(spec string) func() {
	if spec == "" {
		return func() {}
	}
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for _, part := range strings.Split(spec, ",") {
		kind, arg, _ := strings.Cut(part, ":")
		n, _ := strconv.Atoi(arg)
		switch kind {
		case "cpu":
			if n <= 0 {
				n = runtime.NumCPU() / 2
			}
			for i := 0; i < n; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					runtime.LockOSThread()
					x := 0
					for ctx.Err() == nil {
						for k := 0; k < 1_000_000; k++ {
							x += k ^ (x >> 3)
						}
					}
					_ = x
				}()
			}
		case "ram":
			wg.Add(1)
			go func() {
				defer wg.Done()
				holdRAM(ctx, n)
			}()
		case "down", "up":
			wg.Add(1)
			go func(k string) {
				defer wg.Done()
				saturate(ctx, k)
			}(kind)
		}
	}
	return func() {
		cancel()
		wg.Wait()
	}
}

func holdRAM(ctx context.Context, wantMB int) {
	avail := sysmon.AvailableMB()
	mb := float64(wantMB)
	if avail-mb < ramFloorMB {
		mb = avail - ramFloorMB
	}
	if mb < 64 {
		return
	}
	size := uintptr(mb) << 20
	addr, err := windows.VirtualAlloc(0, size, windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE)
	if err != nil {
		return
	}
	defer windows.VirtualFree(addr, 0, windows.MEM_RELEASE)
	buf := unsafe.Slice((*byte)(*(*unsafe.Pointer)(unsafe.Pointer(&addr))), size)
	for ctx.Err() == nil {
		for i := 0; i < len(buf); i += 4096 {
			buf[i]++
		}
		if sysmon.AvailableMB() < ramFloorMB-256 {
			return
		}
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
		}
	}
}

func saturate(ctx context.Context, dir string) {
	client := &http.Client{Timeout: 0}
	for ctx.Err() == nil {
		var req *http.Request
		if dir == "down" {
			req, _ = http.NewRequestWithContext(ctx, http.MethodGet, "https://speed.cloudflare.com/__down?bytes=200000000", nil)
		} else {
			req, _ = http.NewRequestWithContext(ctx, http.MethodPost, "https://speed.cloudflare.com/__up", io.LimitReader(zeroReader{}, 50_000_000))
		}
		resp, err := client.Do(req)
		if err != nil {
			select {
			case <-ctx.Done():
			case <-time.After(time.Second):
			}
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
