package stats

import (
	"math"
	"math/rand"
	"sort"
)

func Mean(x []float64) float64 {
	if len(x) == 0 {
		return math.NaN()
	}
	s := 0.0
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}

func Stdev(x []float64) float64 {
	if len(x) < 2 {
		return 0
	}
	m := Mean(x)
	s := 0.0
	for _, v := range x {
		s += (v - m) * (v - m)
	}
	return math.Sqrt(s / float64(len(x)-1))
}

func Sorted(x []float64) []float64 {
	c := append([]float64(nil), x...)
	sort.Float64s(c)
	return c
}

func Percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return math.NaN()
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	k := (float64(len(sorted)) - 1) * p / 100
	f := math.Floor(k)
	c := math.Min(f+1, float64(len(sorted)-1))
	return sorted[int(f)] + (sorted[int(c)]-sorted[int(f)])*(k-f)
}

func Median(x []float64) float64 {
	return Percentile(Sorted(x), 50)
}

func TopMean(x []float64, fraction float64) float64 {
	s := Sorted(x)
	n := int(math.Max(1, math.Round(float64(len(s))*fraction)))
	if n > len(s) {
		n = len(s)
	}
	return Mean(s[len(s)-n:])
}

type Comparison struct {
	MedianA    float64 `json:"medianA"`
	MedianB    float64 `json:"medianB"`
	DeltaPct   float64 `json:"deltaPct"`
	CILowPct   float64 `json:"ciLowPct"`
	CIHighPct  float64 `json:"ciHighPct"`
	PValue     float64 `json:"pValue"`
	Samples    int     `json:"samples"`
	Conclusive bool    `json:"conclusive"`
}

func CompareAlpha(a, b []float64, seed int64, alpha float64) Comparison {
	c := Compare(a, b, seed)
	c.Conclusive = (c.CILowPct > 0 || c.CIHighPct < 0) && c.PValue < alpha
	return c
}

func Compare(a, b []float64, seed int64) Comparison {
	c := Comparison{MedianA: Median(a), MedianB: Median(b), Samples: len(a) + len(b)}
	if len(a) == 0 || len(b) == 0 || c.MedianA == 0 {
		return c
	}
	c.DeltaPct = 100 * (c.MedianB - c.MedianA) / math.Abs(c.MedianA)
	rng := rand.New(rand.NewSource(seed))
	const rounds = 2000
	deltas := make([]float64, rounds)
	ra := make([]float64, len(a))
	rb := make([]float64, len(b))
	for i := 0; i < rounds; i++ {
		for j := range ra {
			ra[j] = a[rng.Intn(len(a))]
		}
		for j := range rb {
			rb[j] = b[rng.Intn(len(b))]
		}
		ma := Median(ra)
		if ma == 0 {
			continue
		}
		deltas[i] = 100 * (Median(rb) - ma) / math.Abs(ma)
	}
	sort.Float64s(deltas)
	c.CILowPct = Percentile(deltas, 2.5)
	c.CIHighPct = Percentile(deltas, 97.5)
	c.PValue = MannWhitney(a, b)
	c.Conclusive = (c.CILowPct > 0 || c.CIHighPct < 0) && c.PValue < 0.05
	return c
}

func MannWhitney(a, b []float64) float64 {
	type item struct {
		v     float64
		group int
	}
	all := make([]item, 0, len(a)+len(b))
	for _, v := range a {
		all = append(all, item{v, 0})
	}
	for _, v := range b {
		all = append(all, item{v, 1})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].v < all[j].v })
	ranks := make([]float64, len(all))
	tieSum := 0.0
	for i := 0; i < len(all); {
		j := i
		for j < len(all) && all[j].v == all[i].v {
			j++
		}
		r := float64(i+j+1) / 2
		for k := i; k < j; k++ {
			ranks[k] = r
		}
		t := float64(j - i)
		tieSum += t*t*t - t
		i = j
	}
	ra := 0.0
	for i, it := range all {
		if it.group == 0 {
			ra += ranks[i]
		}
	}
	n1, n2 := float64(len(a)), float64(len(b))
	u := ra - n1*(n1+1)/2
	mu := n1 * n2 / 2
	n := n1 + n2
	sigma := math.Sqrt(n1 * n2 / 12 * ((n + 1) - tieSum/(n*(n-1))))
	if sigma == 0 {
		return 1
	}
	z := math.Max(0, math.Abs(u-mu)-0.5) / sigma
	return math.Min(1, math.Erfc(z/math.Sqrt2))
}
