# stats — standard-library-only Go port of the core of the npm package simple-statistics

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/stats.svg)](https://pkg.go.dev/github.com/malcolmston/express/stats)

Package stats is a standard-library-only Go port of the core of the npm
package simple-statistics (https://simplestatistics.org), providing the
descriptive-statistics helpers that data-facing Express/Node services reach
for: measures of central tendency (Mean, Median, Mode and the geometric and
harmonic means), dispersion (Variance, StandardDeviation and their sample
variants, Range, InterquartileRange, MedianAbsoluteDeviation), quantiles
(Quantile, Percentile), bivariate measures (Covariance, Correlation,
LinearRegression) and a few combinatorial functions (Factorial, Combinations,
Permutations).

All sequence functions operate on []float64 and treat the slice as an
unordered sample; those that need order (Median, Quantile and friends) sort a
private copy, so the caller's slice is never modified. Quantile uses linear
interpolation of the empirical CDF (index = p*(n-1)), the same method as
NumPy's default and Percentile is its 0–100 wrapper. Variance and
StandardDeviation are population statistics (divide by n); the Sample*
variants use Bessel's correction (divide by n-1). Covariance and Correlation
are population statistics over paired slices of equal length.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/stats"
```

## Usage

This is the package's own `ExampleAddToMean`, so it compiles and its output is
asserted on every `go test ./stats/`.

```go
mean := 0.0
	for i, x := range []float64{1, 2, 3, 4, 10} {
		mean = stats.AddToMean(mean, i, x)
	}
	fmt.Println(mean)
```

```
4
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func AddToMean(mean float64, n int, newValue float64) float64` | AddToMean returns the mean of a sample of n values whose mean is mean once newValue is appended to it, without revisiting the original values: |
| `func Chunk(xs []float64, chunkSize int) [][]float64` | Chunk splits xs into consecutive slices of length chunkSize, the port of simple-statistics' chunk. |
| `func Combinations(n, k int) float64` | Combinations returns the binomial coefficient C(n, k) as a float64, or NaN when either argument is negative. |
| `func Correlation(xs, ys []float64) float64` | Correlation returns the Pearson correlation coefficient of xs and ys, a value in [-1,1], or NaN when the slices are empty, of unequal length, or… |
| `func Covariance(xs, ys []float64) float64` | Covariance returns the population covariance of the paired slices xs and ys, or NaN when they are empty or of unequal length. |
| `func CumulativeStdNormalProbability(z float64) float64` | CumulativeStdNormalProbability returns P(Z <= z) for a standard normal variable, computed exactly as 0.5*(1+erf(z/sqrt(2))) rather than by the table… |
| `func CumulativeSum(xs []float64) []float64` | CumulativeSum returns a slice whose i-th element is the sum of xs[0..i]. |
| `func ErrorFunction(x float64) float64` | ErrorFunction returns the Gauss error function erf(x), the port of simple-statistics' errorFunction. |
| `func Extent(xs []float64) (min, max float64)` | Extent returns the smallest and largest values in xs in one pass, the port of simple-statistics' extent. |
| `func Factorial(n int) float64` | Factorial returns n! as a float64, or NaN for negative n. |
| `func Frequencies(xs []float64) map[float64]int` | Frequencies returns a map from each distinct value in xs to the number of times it occurs. |
| `func GeometricMean(xs []float64) float64` | GeometricMean returns the geometric mean of xs, or NaN when xs is empty or contains a non-positive value. |
| `func HarmonicMean(xs []float64) float64` | HarmonicMean returns the harmonic mean of xs, or NaN when xs is empty or contains a non-positive value. |
| `func InterquartileRange(xs []float64) float64` | InterquartileRange returns the difference between the 75th and 25th percentiles of xs, or NaN when xs is empty. |
| `func LinearRegression(xs, ys []float64) (slope, intercept float64)` | LinearRegression fits the ordinary least-squares line y = slope*x + intercept to the paired slices and returns its slope and intercept. |
| `func Max(xs []float64) float64` | Max returns the largest value in xs, or NaN when xs is empty. |
| `func Mean(xs []float64) float64` | Mean returns the arithmetic mean of xs, or NaN when xs is empty. |
| `func Median(xs []float64) float64` | Median returns the median of xs (the average of the two middle values for an even count), or NaN when xs is empty. |
| `func MedianAbsoluteDeviation(xs []float64) float64` | MedianAbsoluteDeviation returns the median of the absolute deviations of xs from their median, a robust measure of spread. |
| `func Min(xs []float64) float64` | Min returns the smallest value in xs, or NaN when xs is empty. |
| `func Mode(xs []float64) []float64` | Mode returns the value or values that occur most frequently in xs, sorted ascending. |
| `func Percentile(xs []float64, p float64) float64` | Percentile returns the p-th percentile of xs (p in [0,100]), equivalent to Quantile(xs, p/100). |
| `func Permutations(n, k int) float64` | Permutations returns the number of ordered k-arrangements of n items, n!/(n-k)!, as a float64, or NaN when either argument is negative. |
| `func Product(xs []float64) float64` | Product returns the product of xs. |
| `func Quantile(xs []float64, p float64) float64` | Quantile returns the p-quantile of xs (p in [0,1]) using linear interpolation of the empirical CDF. |
| `func Range(xs []float64) float64` | Range returns the difference between the largest and smallest values in xs, or NaN when xs is empty. |
| `func RootMeanSquare(xs []float64) float64` | RootMeanSquare returns the quadratic mean (root mean square) of xs, or NaN when xs is empty. |
| `func SampleCorrelation(xs, ys []float64) float64` | SampleCorrelation returns the Pearson correlation coefficient of xs and ys computed from the sample covariance and sample standard deviations, the… |
| `func SampleCovariance(xs, ys []float64) float64` | SampleCovariance returns the Bessel-corrected covariance of the paired slices xs and ys (dividing by n-1), the port of simple-statistics'… |
| `func SampleKurtosis(xs []float64) float64` | SampleKurtosis returns the excess kurtosis of xs, |
| `func SampleSkewness(xs []float64) float64` | SampleSkewness returns the adjusted Fisher-Pearson standardized moment coefficient G1 of xs, |
| `func SampleStandardDeviation(xs []float64) float64` | SampleStandardDeviation returns the sample standard deviation of xs, or NaN when xs has fewer than two elements. |
| `func SampleVariance(xs []float64) float64` | SampleVariance returns the sample variance of xs (Bessel-corrected, dividing by n-1), or NaN when xs has fewer than two elements. |
| `func StandardDeviation(xs []float64) float64` | StandardDeviation returns the population standard deviation of xs, or NaN when xs is empty. |
| `func Sum(xs []float64) float64` | Sum returns the sum of xs. |
| `func SumNthPowerDeviations(xs []float64, n int) float64` | SumNthPowerDeviations returns the sum of (x - mean(xs))^n over xs, the raw n-th central moment sum that skewness and kurtosis are built from. |
| `func UniqueCountSorted(sorted []float64) int` | UniqueCountSorted returns the number of distinct values in the already-sorted slice sorted, the port of simple-statistics' uniqueCountSorted. |
| `func Variance(xs []float64) float64` | Variance returns the population variance of xs (dividing by n), or NaN when xs is empty. |
| `func ZScore(x, mean, standardDeviation float64) float64` | ZScore returns the number of standard deviations x lies from a distribution with the given mean and standard deviation. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/stats).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
