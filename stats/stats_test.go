package stats

import (
	"math"
	"reflect"
	"testing"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestBasics(t *testing.T) {
	xs := []float64{1, 2, 3, 4, 5}
	if Sum(xs) != 15 {
		t.Error("Sum")
	}
	if Product([]float64{1, 2, 3, 4}) != 24 {
		t.Error("Product")
	}
	if Mean(xs) != 3 {
		t.Error("Mean")
	}
	if Min(xs) != 1 || Max(xs) != 5 || Range(xs) != 4 {
		t.Error("Min/Max/Range")
	}
	if !math.IsNaN(Mean(nil)) {
		t.Error("Mean empty should be NaN")
	}
}

func TestMeans(t *testing.T) {
	if !approx(GeometricMean([]float64{1, 3, 9}), 3) {
		t.Errorf("GeometricMean = %g", GeometricMean([]float64{1, 3, 9}))
	}
	if !approx(HarmonicMean([]float64{1, 2, 4}), 12.0/7.0) {
		t.Errorf("HarmonicMean = %g", HarmonicMean([]float64{1, 2, 4}))
	}
	if !approx(RootMeanSquare([]float64{3, 4}), math.Sqrt(12.5)) {
		t.Error("RootMeanSquare")
	}
}

func TestMedianMode(t *testing.T) {
	if Median([]float64{1, 2, 3, 4, 5}) != 3 {
		t.Error("Median odd")
	}
	if Median([]float64{1, 2, 3, 4}) != 2.5 {
		t.Error("Median even")
	}
	if got := Mode([]float64{1, 2, 2, 3, 3, 4}); !reflect.DeepEqual(got, []float64{2, 3}) {
		t.Errorf("Mode = %v", got)
	}
	if got := Mode([]float64{5, 1, 1, 5, 5}); !reflect.DeepEqual(got, []float64{5}) {
		t.Errorf("Mode single = %v", got)
	}
}

func TestQuantile(t *testing.T) {
	xs := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if !approx(Quantile(xs, 0.5), 5.5) {
		t.Errorf("Quantile 0.5 = %g", Quantile(xs, 0.5))
	}
	if !approx(Percentile(xs, 25), 3.25) {
		t.Errorf("Percentile 25 = %g", Percentile(xs, 25))
	}
	if !approx(InterquartileRange(xs), 4.5) {
		t.Errorf("IQR = %g", InterquartileRange(xs))
	}
	if !math.IsNaN(Quantile(xs, 1.5)) {
		t.Error("Quantile out of range")
	}
}

func TestVariance(t *testing.T) {
	xs := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	if !approx(Variance(xs), 4) {
		t.Errorf("Variance = %g", Variance(xs))
	}
	if !approx(StandardDeviation(xs), 2) {
		t.Errorf("StandardDeviation = %g", StandardDeviation(xs))
	}
	// sample variance of [1,2,3,4,5] = 2.5
	if !approx(SampleVariance([]float64{1, 2, 3, 4, 5}), 2.5) {
		t.Errorf("SampleVariance = %g", SampleVariance([]float64{1, 2, 3, 4, 5}))
	}
	if !approx(SampleStandardDeviation([]float64{1, 2, 3, 4, 5}), math.Sqrt(2.5)) {
		t.Error("SampleStandardDeviation")
	}
	if !math.IsNaN(SampleVariance([]float64{1})) {
		t.Error("SampleVariance n<2")
	}
}

func TestMAD(t *testing.T) {
	if !approx(MedianAbsoluteDeviation([]float64{1, 1, 2, 2, 4, 6, 9}), 1) {
		t.Errorf("MAD = %g", MedianAbsoluteDeviation([]float64{1, 1, 2, 2, 4, 6, 9}))
	}
}

func TestZScore(t *testing.T) {
	if ZScore(85, 75, 5) != 2 {
		t.Error("ZScore")
	}
}

func TestBivariate(t *testing.T) {
	xs := []float64{1, 2, 3, 4, 5}
	ys := []float64{2, 4, 6, 8, 10}
	if !approx(Correlation(xs, ys), 1) {
		t.Errorf("Correlation = %g", Correlation(xs, ys))
	}
	if !approx(Covariance(xs, ys), 4) {
		t.Errorf("Covariance = %g", Covariance(xs, ys))
	}
	slope, intercept := LinearRegression(xs, ys)
	if !approx(slope, 2) || !approx(intercept, 0) {
		t.Errorf("LinearRegression = %g, %g", slope, intercept)
	}
	if !math.IsNaN(Correlation(xs, []float64{1, 2})) {
		t.Error("mismatched length")
	}
}

func TestCumulativeSum(t *testing.T) {
	got := CumulativeSum([]float64{1, 2, 3, 4})
	if !reflect.DeepEqual(got, []float64{1, 3, 6, 10}) {
		t.Errorf("CumulativeSum = %v", got)
	}
	if len(CumulativeSum(nil)) != 0 {
		t.Error("CumulativeSum empty")
	}
}

func TestFrequencies(t *testing.T) {
	got := Frequencies([]float64{1, 1, 2, 3, 3, 3})
	want := map[float64]int{1: 2, 2: 1, 3: 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Frequencies = %v", got)
	}
}

func TestCombinatorics(t *testing.T) {
	if Factorial(5) != 120 || Factorial(0) != 1 {
		t.Error("Factorial")
	}
	if !math.IsNaN(Factorial(-1)) {
		t.Error("Factorial neg")
	}
	if Combinations(5, 2) != 10 {
		t.Errorf("Combinations = %g", Combinations(5, 2))
	}
	if Combinations(5, 6) != 0 {
		t.Error("Combinations k>n")
	}
	if Permutations(5, 2) != 20 {
		t.Errorf("Permutations = %g", Permutations(5, 2))
	}
}

func BenchmarkQuantile(b *testing.B) {
	xs := make([]float64, 100)
	for i := range xs {
		xs[i] = float64(100 - i)
	}
	for i := 0; i < b.N; i++ {
		_ = Quantile(xs, 0.9)
	}
}
