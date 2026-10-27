package netq

import (
	"encoding/binary"
	"errors"
	"math"
	"net/netip"
	"os"
)

type Packet struct {
	T     float64
	Out   bool
	Len   int
	Seq   uint16
	HasSq bool
}

const (
	inboundAckLen  = 54
	inboundAckLen2 = 57
	outboundAckLen  = 56
	outboundAckLen2 = 59
	seqOffset       = 36
)

func ReadPcapng(path string, server netip.Addr) ([]Packet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	srv := server.As4()
	var out []Packet
	var tsUnit = []float64{}
	le := binary.LittleEndian
	for off := 0; off+12 <= len(data); {
		btype := le.Uint32(data[off:])
		blen := int(le.Uint32(data[off+4:]))
		if blen < 12 || off+blen > len(data) {
			break
		}
		body := data[off+8 : off+blen-4]
		switch btype {
		case 0x00000001:
			tsUnit = append(tsUnit, idbResolution(body))
		case 0x00000006:
			if len(body) < 20 {
				break
			}
			iface := int(le.Uint32(body[0:]))
			ts := uint64(le.Uint32(body[4:]))<<32 | uint64(le.Uint32(body[8:]))
			capLen := int(le.Uint32(body[12:]))
			if 20+capLen > len(body) {
				capLen = len(body) - 20
			}
			unit := 1e-6
			if iface < len(tsUnit) {
				unit = tsUnit[iface]
			}
			if p, ok := parseFrame(body[20:20+capLen], srv); ok {
				p.T = float64(ts) * unit
				out = append(out, p)
			}
		}
		off += blen
	}
	if len(out) == 0 {
		return nil, errors.New("no roblox packets captured")
	}
	return out, nil
}

func idbResolution(body []byte) float64 {
	unit := 1e-6
	for o := 8; o+4 <= len(body); {
		code := binary.LittleEndian.Uint16(body[o:])
		l := int(binary.LittleEndian.Uint16(body[o+2:]))
		if code == 0 {
			break
		}
		if code == 9 && l >= 1 && o+4 < len(body) {
			v := body[o+4]
			if v&0x80 == 0 {
				unit = math.Pow(10, -float64(v))
			} else {
				unit = math.Pow(2, -float64(v&0x7f))
			}
		}
		o += 4 + (l+3)&^3
	}
	return unit
}

func parseFrame(f []byte, srv [4]byte) (Packet, bool) {
	if len(f) < 14+20+8 {
		return Packet{}, false
	}
	o := 14
	if binary.BigEndian.Uint16(f[12:]) == 0x8100 {
		o = 18
	}
	if len(f) < o+20 || f[o]>>4 != 4 || f[o+9] != 17 {
		return Packet{}, false
	}
	ihl := int(f[o]&0x0f) * 4
	var src, dst [4]byte
	copy(src[:], f[o+12:o+16])
	copy(dst[:], f[o+16:o+20])
	u := o + ihl
	if len(f) < u+8 {
		return Packet{}, false
	}
	udpLen := int(binary.BigEndian.Uint16(f[u+4:])) - 8
	p := Packet{Len: udpLen}
	switch {
	case dst == srv:
		p.Out = true
	case src == srv:
		p.Out = false
	default:
		return Packet{}, false
	}
	payload := f[u+8:]
	if !p.Out && udpLen == inboundAckLen && len(payload) >= seqOffset+2 {
		p.Seq = binary.LittleEndian.Uint16(payload[seqOffset:])
		p.HasSq = true
	}
	return p, true
}
