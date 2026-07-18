// Package stats is a standard-library-only Go port of the core of the npm
// package simple-statistics (https://simplestatistics.org), providing the
// descriptive-statistics helpers that data-facing Express/Node services reach
// for: measures of central tendency (Mean, Median, Mode and the geometric and
// harmonic means), dispersion (Variance, StandardDeviation and their sample
// variants, Range, InterquartileRange, MedianAbsoluteDeviation), quantiles
// (Quantile, Percentile), bivariate measures (Covariance, Correlation,
// LinearRegression) and a few combinatorial functions (Factorial, Combinations,
// Permutations).
//
// All sequence functions operate on []float64 and treat the slice as an
// unordered sample; those that need order (Median, Quantile and friends) sort a
// private copy, so the caller's slice is never modified. Quantile uses linear
// interpolation of the empirical CDF (index = p*(n-1)), the same method as
// NumPy's default and Percentile is its 0–100 wrapper. Variance and
// StandardDeviation are population statistics (divide by n); the Sample*
// variants use Bessel's correction (divide by n-1). Covariance and Correlation
// are population statistics over paired slices of equal length.
//
// Functions return NaN for inputs that leave the statistic undefined — an empty
// slice for Mean, fewer than two elements for a sample variance, or mismatched
// lengths for the bivariate helpers — mirroring simple-statistics returning
// undefined, but in a form Go callers can test with math.IsNaN. Everything is
// deterministic and depends only on the math and sort packages.
package stats

import (
	"math"
	"sort"
)

// Sum returns the sum of xs. The sum of an empty slice is 0.
func Sum(xs []float64) float64 {
	var s float64
	for _, x := range xs {
		s += x
	}
	return s
}

// Product returns the product of xs. The product of an empty slice is 1.
func Product(xs []float64) float64 {
	p := 1.0
	for _, x := range xs {
		p *= x
	}
	return p
}

// Mean returns the arithmetic mean of xs, or NaN when xs is empty.
func Mean(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	return Sum(xs) / float64(len(xs))
}

// GeometricMean returns the geometric mean of xs, or NaN when xs is empty or
// contains a non-positive value.
func GeometricMean(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	sumLog := 0.0
	for _, x := range xs {
		if x <= 0 {
			return math.NaN()
		}
		sumLog += math.Log(x)
	}
	return math.Exp(sumLog / float64(len(xs)))
}

// HarmonicMean returns the harmonic mean of xs, or NaN when xs is empty or
// contains a non-positive value.
func HarmonicMean(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	var recip float64
	for _, x := range xs {
		if x <= 0 {
			return math.NaN()
		}
		recip += 1 / x
	}
	return float64(len(xs)) / recip
}

// Min returns the smallest value in xs, or NaN when xs is empty.
func Min(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	m := xs[0]
	for _, x := range xs[1:] {
		if x < m {
			m = x
		}
	}
	return m
}

// Max returns the largest value in xs, or NaN when xs is empty.
func Max(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	m := xs[0]
	for _, x := range xs[1:] {
		if x > m {
			m = x
		}
	}
	return m
}

// Range returns the difference between the largest and smallest values in xs,
// or NaN when xs is empty.
func Range(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	return Max(xs) - Min(xs)
}

func statsSorted(xs []float64) []float64 {
	c := make([]float64, len(xs))
	copy(c, xs)
	sort.Float64s(c)
	return c
}

// Median returns the median of xs (the average of the two middle values for an
// even count), or NaN when xs is empty.
func Median(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	s := statsSorted(xs)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

// Mode returns the value or values that occur most frequently in xs, sorted
// ascending. It returns an empty slice when xs is empty.
func Mode(xs []float64) []float64 {
	out := []float64{}
	if len(xs) == 0 {
		return out
	}
	s := statsSorted(xs)
	best := 0
	counts := map[float64]int{}
	for _, x := range s {
		counts[x]++
		if counts[x] > best {
			best = counts[x]
		}
	}
	seen := map[float64]bool{}
	for _, x := range s {
		if counts[x] == best && !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}

// Quantile returns the p-quantile of xs (p in [0,1]) using linear interpolation
// of the empirical CDF. Quantile(xs, 0.5) equals Median(xs). It returns NaN for
// an empty slice or a p outside [0,1].
func Quantile(xs []float64, p float64) float64 {
	if len(xs) == 0 || p < 0 || p > 1 {
		return math.NaN()
	}
	s := statsSorted(xs)
	if len(s) == 1 {
		return s[0]
	}
	pos := p * float64(len(s)-1)
	lo := int(math.Floor(pos))
	frac := pos - float64(lo)
	if lo+1 >= len(s) {
		return s[lo]
	}
	return s[lo] + frac*(s[lo+1]-s[lo])
}

// Percentile returns the p-th percentile of xs (p in [0,100]), equivalent to
// Quantile(xs, p/100).
func Percentile(xs []float64, p float64) float64 {
	return Quantile(xs, p/100)
}

// InterquartileRange returns the difference between the 75th and 25th
// percentiles of xs, or NaN when xs is empty.
func InterquartileRange(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	return Quantile(xs, 0.75) - Quantile(xs, 0.25)
}

// Variance returns the population variance of xs (dividing by n), or NaN when
// xs is empty.
func Variance(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	m := Mean(xs)
	var ss float64
	for _, x := range xs {
		d := x - m
		ss += d * d
	}
	return ss / float64(len(xs))
}

// SampleVariance returns the sample variance of xs (Bessel-corrected, dividing
// by n-1), or NaN when xs has fewer than two elements.
func SampleVariance(xs []float64) float64 {
	if len(xs) < 2 {
		return math.NaN()
	}
	m := Mean(xs)
	var ss float64
	for _, x := range xs {
		d := x - m
		ss += d * d
	}
	return ss / float64(len(xs)-1)
}

// StandardDeviation returns the population standard deviation of xs, or NaN when
// xs is empty.
func StandardDeviation(xs []float64) float64 {
	return math.Sqrt(Variance(xs))
}

// SampleStandardDeviation returns the sample standard deviation of xs, or NaN
// when xs has fewer than two elements.
func SampleStandardDeviation(xs []float64) float64 {
	return math.Sqrt(SampleVariance(xs))
}

// RootMeanSquare returns the quadratic mean (root mean square) of xs, or NaN
// when xs is empty.
func RootMeanSquare(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	var ss float64
	for _, x := range xs {
		ss += x * x
	}
	return math.Sqrt(ss / float64(len(xs)))
}

// MedianAbsoluteDeviation returns the median of the absolute deviations of xs
// from their median, a robust measure of spread. It returns NaN for an empty
// slice.
func MedianAbsoluteDeviation(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	med := Median(xs)
	dev := make([]float64, len(xs))
	for i, x := range xs {
		dev[i] = math.Abs(x - med)
	}
	return Median(dev)
}

// ZScore returns the number of standard deviations x lies from a distribution
// with the given mean and standard deviation.
func ZScore(x, mean, standardDeviation float64) float64 {
	return (x - mean) / standardDeviation
}

// Covariance returns the population covariance of the paired slices xs and ys,
// or NaN when they are empty or of unequal length.
func Covariance(xs, ys []float64) float64 {
	if len(xs) != len(ys) || len(xs) == 0 {
		return math.NaN()
	}
	mx, my := Mean(xs), Mean(ys)
	var s float64
	for i := range xs {
		s += (xs[i] - mx) * (ys[i] - my)
	}
	return s / float64(len(xs))
}

// Correlation returns the Pearson correlation coefficient of xs and ys, a value
// in [-1,1], or NaN when the slices are empty, of unequal length, or either has
// zero variance.
func Correlation(xs, ys []float64) float64 {
	if len(xs) != len(ys) || len(xs) == 0 {
		return math.NaN()
	}
	sx := StandardDeviation(xs)
	sy := StandardDeviation(ys)
	if sx == 0 || sy == 0 {
		return math.NaN()
	}
	return Covariance(xs, ys) / (sx * sy)
}

// LinearRegression fits the ordinary least-squares line y = slope*x + intercept
// to the paired slices and returns its slope and intercept. It returns NaN
// values when the slices are empty, of unequal length, or x has zero variance.
func LinearRegression(xs, ys []float64) (slope, intercept float64) {
	if len(xs) != len(ys) || len(xs) == 0 {
		return math.NaN(), math.NaN()
	}
	mx, my := Mean(xs), Mean(ys)
	var num, den float64
	for i := range xs {
		dx := xs[i] - mx
		num += dx * (ys[i] - my)
		den += dx * dx
	}
	if den == 0 {
		return math.NaN(), math.NaN()
	}
	slope = num / den
	intercept = my - slope*mx
	return slope, intercept
}

// CumulativeSum returns a slice whose i-th element is the sum of xs[0..i]. The
// result has the same length as xs; an empty input yields an empty slice.
func CumulativeSum(xs []float64) []float64 {
	out := make([]float64, len(xs))
	var acc float64
	for i, x := range xs {
		acc += x
		out[i] = acc
	}
	return out
}

// Frequencies returns a map from each distinct value in xs to the number of
// times it occurs.
func Frequencies(xs []float64) map[float64]int {
	m := make(map[float64]int)
	for _, x := range xs {
		m[x]++
	}
	return m
}

// Factorial returns n! as a float64, or NaN for negative n. Factorial(0) is 1.
func Factorial(n int) float64 {
	if n < 0 {
		return math.NaN()
	}
	r := 1.0
	for i := 2; i <= n; i++ {
		r *= float64(i)
	}
	return r
}

// Combinations returns the binomial coefficient C(n, k) as a float64, or NaN
// when either argument is negative. It is 0 when k > n.
func Combinations(n, k int) float64 {
	if n < 0 || k < 0 {
		return math.NaN()
	}
	if k > n {
		return 0
	}
	if k > n-k {
		k = n - k
	}
	r := 1.0
	for i := 0; i < k; i++ {
		r = r * float64(n-i) / float64(i+1)
	}
	return math.Round(r)
}

// Permutations returns the number of ordered k-arrangements of n items,
// n!/(n-k)!, as a float64, or NaN when either argument is negative. It is 0
// when k > n.
func Permutations(n, k int) float64 {
	if n < 0 || k < 0 {
		return math.NaN()
	}
	if k > n {
		return 0
	}
	r := 1.0
	for i := 0; i < k; i++ {
		r *= float64(n - i)
	}
	return r
}
