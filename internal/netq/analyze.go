package netq

import (
	"math"
	"sort"

	"toque/internal/stats"
)

const ServerAckDelayMs = 10.3

type Summary struct {
	Packets      int       `json:"packets"`
	Seconds      float64   `json:"seconds"`
	PPSIn        float64   `json:"ppsIn"`
	PPSOut       float64   `json:"ppsOut"`
	KbpsIn       float64   `json:"kbpsIn"`
	KbpsOut      float64   `json:"kbpsOut"`
	RTT          float64   `json:"rttMs"`
	RTTP95       float64   `json:"rttP95Ms"`
	Jitter       float64   `json:"jitterMs"`
	AckPeak      float64   `json:"ackPeakMs"`
	Matched      int       `json:"matched"`
	ExpectedDown int       `json:"expectedDown"`
	LostDown     int       `json:"lostDown"`
	LossDownPct  float64   `json:"lossDownPct"`
	MaxGapMs     float64   `json:"maxGapMs"`
	GapP99Ms     float64   `json:"gapP99Ms"`
	Stalls100    int       `json:"stalls100"`
	Stalls250    int       `json:"stalls250"`
	Series       []float64 `json:"rttSeries"`
	Confident    bool      `json:"confident"`
}

func Analyze(pk []Packet) Summary {
	var s Summary
	if len(pk) < 2 {
		return s
	}
	sort.Slice(pk, func(i, j int) bool { return pk[i].T < pk[j].T })
	s.Packets = len(pk)
	s.Seconds = pk[len(pk)-1].T - pk[0].T
	var sends, acks, inbound []float64
	var bytesIn, bytesOut int
	for _, p := range pk {
		if p.Out {
			bytesOut += p.Len
			if p.Len != outboundAckLen && p.Len != outboundAckLen2 {
				sends = append(sends, p.T)
			}
		} else {
			bytesIn += p.Len
			inbound = append(inbound, p.T)
			if p.Len == inboundAckLen || p.Len == inboundAckLen2 {
				acks = append(acks, p.T)
			}
		}
	}
	if s.Seconds > 0 {
		s.PPSIn = float64(len(inbound)) / s.Seconds
		s.PPSOut = float64(s.Packets-len(inbound)) / s.Seconds
		s.KbpsIn = float64(bytesIn) * 8 / 1000 / s.Seconds
		s.KbpsOut = float64(bytesOut) * 8 / 1000 / s.Seconds
	}
	peak, samples := estimateRTT(sends, acks)
	if len(samples) > 0 {
		s.AckPeak = stats.Median(samples)
		s.RTT = s.AckPeak - ServerAckDelayMs
		srt := stats.Sorted(samples)
		s.RTTP95 = stats.Percentile(srt, 95) - ServerAckDelayMs
		s.Jitter = stats.Stdev(samples)
		s.Matched = len(samples)
		s.Confident = len(samples) >= 100 && peak > 0
	}
	s.Series = rttSeries(sends, acks, 5)
	s.ExpectedDown, s.LostDown = downstreamLoss(pk)
	if s.ExpectedDown > 0 {
		s.LossDownPct = 100 * float64(s.LostDown) / float64(s.ExpectedDown)
	}
	var gaps []float64
	for i := 1; i < len(inbound); i++ {
		g := (inbound[i] - inbound[i-1]) * 1000
		gaps = append(gaps, g)
		if g > 100 {
			s.Stalls100++
		}
		if g > 250 {
			s.Stalls250++
		}
	}
	if len(gaps) > 0 {
		sg := stats.Sorted(gaps)
		s.MaxGapMs = sg[len(sg)-1]
		s.GapP99Ms = stats.Percentile(sg, 99)
	}
	return s
}

func estimateRTT(sends, acks []float64) (float64, []float64) {
	const bins = 800
	hist := make([]float64, bins)
	for _, a := range acks {
		lo := sort.SearchFloat64s(sends, a-0.400)
		hi := sort.SearchFloat64s(sends, a-0.005)
		for _, snd := range sends[lo:hi] {
			b := int((a - snd) * 2000)
			if b >= 0 && b < bins {
				hist[b]++
			}
		}
	}
	best, bestScore := -1, 0.0
	for i := 20; i < bins-41; i++ {
		var window []float64
		window = append(window, hist[max(0, i-40):i-6]...)
		window = append(window, hist[i+7:i+40]...)
		base := stats.Median(window)
		score := hist[i-1] + hist[i] + hist[i+1] - 3*base
		if score > bestScore {
			best, bestScore = i, score
		}
	}
	if best < 0 {
		return 0, nil
	}
	center := float64(best) / 2000
	var samples []float64
	for _, a := range acks {
		target := a - center
		j := sort.SearchFloat64s(sends, target)
		cand := math.Inf(1)
		for _, k := range []int{j - 1, j} {
			if k >= 0 && k < len(sends) && math.Abs(sends[k]-target) < math.Abs(cand-target) {
				cand = sends[k]
			}
		}
		if math.IsInf(cand, 1) {
			continue
		}
		d := a - cand
		if math.Abs(d-center) <= 0.006 {
			samples = append(samples, d*1000)
		}
	}
	return center * 1000, samples
}

func rttSeries(sends, acks []float64, windowSecs float64) []float64 {
	if len(acks) == 0 {
		return nil
	}
	var out []float64
	start := acks[0]
	for t := start; t < acks[len(acks)-1]; t += windowSecs {
		lo := sort.SearchFloat64s(acks, t)
		hi := sort.SearchFloat64s(acks, t+windowSecs)
		slo := sort.SearchFloat64s(sends, t-0.5)
		shi := sort.SearchFloat64s(sends, t+windowSecs)
		_, samples := estimateRTT(sends[slo:shi], acks[lo:hi])
		if len(samples) >= 20 {
			out = append(out, stats.Median(samples)-ServerAckDelayMs)
		}
	}
	return out
}

func downstreamLoss(pk []Packet) (int, int) {
	expected, lost := 0, 0
	lastIdx, lastSeq, have := 0, uint16(0), false
	idx := 0
	for _, p := range pk {
		if p.Out {
			continue
		}
		if p.HasSq {
			if have {
				e := int(uint16(p.Seq - lastSeq))
				got := idx - lastIdx
				if e < 2000 {
					expected += e
					if e > got {
						lost += e - got
					}
				}
			}
			lastIdx, lastSeq, have = idx, p.Seq, true
		}
		idx++
	}
	return expected, lost
}
