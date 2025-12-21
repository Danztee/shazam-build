package audio

import (
	"math"
	"math/cmplx"
)

func FFT(samples []float64) []complex128 {
	if len(samples) == 1 {
		return []complex128{complex(samples[0], 0)}
	}

	nextPow2 := 1
	for nextPow2 < len(samples) {
		nextPow2 <<= 1
	}

	padded := make([]float64, nextPow2)
	copy(padded, samples)

	return fftRecursive(padded)
}

func fftRecursive(x []float64) []complex128 {
	n := len(x)
	if n <= 1 {
		result := make([]complex128, n)
		for i := range x {
			result[i] = complex(x[i], 0)
		}
		return result
	}

	even := make([]float64, n/2)
	odd := make([]float64, n/2)
	for i := 0; i < n/2; i++ {
		even[i] = x[i*2]
		odd[i] = x[i*2+1]
	}

	evenFFT := fftRecursive(even)
	oddFFT := fftRecursive(odd)

	result := make([]complex128, n)
	for k := 0; k < n/2; k++ {
		t := cmplx.Exp(complex(0, -2*math.Pi*float64(k)/float64(n))) * oddFFT[k]
		result[k] = evenFFT[k] + t
		result[k+n/2] = evenFFT[k] - t
	}

	return result
}

func MagnitudeSpectrum(fft []complex128) []float64 {
	magnitude := make([]float64, len(fft)/2)
	for i := 0; i < len(magnitude); i++ {
		r, im := real(fft[i]), imag(fft[i])
		magnitude[i] = math.Sqrt(r*r + im*im)
	}
	return magnitude
}
