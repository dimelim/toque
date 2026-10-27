package netq

import (
	"errors"
	"net/netip"
	"sort"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"toque/internal/stats"
)

var (
	iphlpapi         = windows.NewLazySystemDLL("iphlpapi.dll")
	procIcmpCreate   = iphlpapi.NewProc("IcmpCreateFile")
	procIcmpClose    = iphlpapi.NewProc("IcmpCloseHandle")
	procIcmpSendEcho = iphlpapi.NewProc("IcmpSendEcho")
)

type ipOptionInfo struct {
	TTL         uint8
	TOS         uint8
	Flags       uint8
	OptionsSize uint8
	OptionsData uintptr
}

type icmpEchoReply struct {
	Address       uint32
	Status        uint32
	RoundTripTime uint32
	DataSize      uint16
	Reserved      uint16
	Data          uintptr
	Options       ipOptionInfo
}

const (
	ipSuccess     = 0
	ipTTLExpired  = 11013
	ipReqTimedOut = 11010
)

type Echo struct {
	From    netip.Addr
	RTT     time.Duration
	Reached bool
	TTLHop  bool
}

func echo(dst netip.Addr, ttl uint8, timeout time.Duration) (Echo, error) {
	h, _, err := procIcmpCreate.Call()
	if h == uintptr(windows.InvalidHandle) {
		return Echo{}, err
	}
	defer procIcmpClose.Call(h)
	a := dst.As4()
	addr := uint32(a[0]) | uint32(a[1])<<8 | uint32(a[2])<<16 | uint32(a[3])<<24
	payload := []byte("toque-probe-0000000000000000")
	opt := ipOptionInfo{TTL: ttl}
	reply := make([]byte, int(unsafe.Sizeof(icmpEchoReply{}))+len(payload)+8+64)
	start := time.Now()
	n, _, _ := procIcmpSendEcho.Call(h, uintptr(addr),
		uintptr(unsafe.Pointer(&payload[0])), uintptr(len(payload)),
		uintptr(unsafe.Pointer(&opt)),
		uintptr(unsafe.Pointer(&reply[0])), uintptr(len(reply)),
		uintptr(timeout.Milliseconds()))
	elapsed := time.Since(start)
	r := (*icmpEchoReply)(unsafe.Pointer(&reply[0]))
	if n == 0 && r.Status != ipTTLExpired {
		return Echo{}, errors.New("no reply")
	}
	from := netip.AddrFrom4([4]byte{byte(r.Address), byte(r.Address >> 8), byte(r.Address >> 16), byte(r.Address >> 24)})
	rtt := time.Duration(r.RoundTripTime) * time.Millisecond
	if rtt == 0 {
		rtt = elapsed
	}
	return Echo{From: from, RTT: rtt, Reached: r.Status == ipSuccess, TTLHop: r.Status == ipTTLExpired}, nil
}

type Hop struct {
	TTL     int       `json:"ttl"`
	Addr    string    `json:"addr"`
	Private bool      `json:"private"`
	RTT     float64   `json:"rttMs"`
	Jitter  float64   `json:"jitterMs"`
	Loss    float64   `json:"lossPct"`
	Samples []float64 `json:"-"`
}

func Trace(dst netip.Addr, maxHops int, probes int) []Hop {
	var hops []Hop
	silent := 0
	for ttl := 1; ttl <= maxHops; ttl++ {
		hop := Hop{TTL: ttl}
		lost := 0
		for p := 0; p < probes; p++ {
			e, err := echo(dst, uint8(ttl), 900*time.Millisecond)
			if err != nil {
				lost++
				continue
			}
			hop.Addr = e.From.String()
			hop.Private = e.From.IsPrivate() || e.From.IsLoopback()
			hop.Samples = append(hop.Samples, float64(e.RTT.Microseconds())/1000)
			if e.Reached {
				ttl = maxHops
			}
		}
		if len(hop.Samples) == 0 {
			silent++
			if silent >= 2 {
				break
			}
			continue
		}
		silent = 0
		hop.Loss = 100 * float64(lost) / float64(probes)
		hop.RTT = stats.Median(hop.Samples)
		hop.Jitter = stats.Stdev(hop.Samples)
		hops = append(hops, hop)
	}
	return hops
}

func Ping(dst netip.Addr, count int, interval time.Duration) Hop {
	hop := Hop{Addr: dst.String(), Private: dst.IsPrivate()}
	lost := 0
	for i := 0; i < count; i++ {
		e, err := echo(dst, 128, time.Second)
		if err != nil || !e.Reached {
			lost++
		} else {
			hop.Samples = append(hop.Samples, float64(e.RTT.Microseconds())/1000)
		}
		time.Sleep(interval)
	}
	hop.Loss = 100 * float64(lost) / float64(count)
	if len(hop.Samples) > 0 {
		s := append([]float64(nil), hop.Samples...)
		sort.Float64s(s)
		hop.RTT = stats.Percentile(s, 50)
		hop.Jitter = stats.Stdev(hop.Samples)
	}
	return hop
}
