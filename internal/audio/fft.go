package audio

import (
	"math"
	"math/cmplx"
)

func FFT(samples []float64) []complex128 {
	n := len(samples)

	nextPow2 := 1
	for nextPow2 < n {
		nextPow2 <<= 1
	}

	x := make([]complex128, nextPow2)
	for i, s := range samples {
		x[i] = complex(s, 0)
	}

	// Bit-reversal permutation
	j := 0
	for i := 0; i < nextPow2-1; i++ {
		if i < j {
			x[i], x[j] = x[j], x[i]
		}
		k := nextPow2 >> 1
		for k <= j {
			j -= k
			k >>= 1
		}
		j += k
	}

	for ip := 1; ip < nextPow2; ip <<= 1 { // ip is the size of the sub-problem
		ang := -math.Pi / float64(ip)
		wStep := cmplx.Exp(complex(0, ang))

		for i := 0; i < nextPow2; i += 2 * ip {
			w := complex(1, 0)
			for k := 0; k < ip; k++ {
				t := w * x[i+k+ip]
				x[i+k+ip] = x[i+k] - t
				x[i+k] = x[i+k] + t
				w = w * wStep
			}
		}
	}

	return x
}

func MagnitudeSpectrum(fft []complex128) []float64 {
	magnitude := make([]float64, len(fft)/2)
	for i := 0; i < len(magnitude); i++ {
		r, im := real(fft[i]), imag(fft[i])
		magnitude[i] = math.Sqrt(r*r + im*im)
	}
	return magnitude
}
