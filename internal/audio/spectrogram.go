package audio

import "math"

const (
	DefaultWindowSize = 4096
	DefaultHopSize    = 2048
)

func ComputeSpectrogram(audio *AudioData, windowSize, hopSize int) *Spectrogram {
	floatSamples := make([]float64, len(audio.Samples))
	for i, s := range audio.Samples {
		floatSamples[i] = float64(s) / 32768.0
	}

	window := make([]float64, windowSize)
	for i := 0; i < windowSize; i++ {
		window[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(windowSize-1)))
	}

	numWindows := (len(floatSamples)-windowSize)/hopSize + 1
	spectrogram := make([][]float32, numWindows)
	timeBins := make([]float64, numWindows)
	freqBins := make([]float64, windowSize/2)

	for i := 0; i < len(freqBins); i++ {
		freqBins[i] = float64(i) * float64(audio.SampleRate) / float64(windowSize)
	}

	for i := 0; i < numWindows; i++ {
		start := i * hopSize
		if start+windowSize > len(floatSamples) {
			break
		}

		windowed := make([]float64, windowSize)
		for j := 0; j < windowSize; j++ {
			windowed[j] = floatSamples[start+j] * window[j]
		}

		fft := FFT(windowed)
		magnitude := MagnitudeSpectrum(fft)

		magnitudeDB := make([]float32, len(magnitude))
		for j := range magnitude {
			if magnitude[j] > 0 {
				magnitudeDB[j] = float32(20 * math.Log10(magnitude[j]))
			} else {
				magnitudeDB[j] = -100
			}
		}

		spectrogram[i] = magnitudeDB
		timeBins[i] = float64(start) / float64(audio.SampleRate)
	}

	return &Spectrogram{
		Data:          spectrogram,
		TimeBins:      timeBins,
		FrequencyBins: freqBins,
		SampleRate:    audio.SampleRate,
		WindowSize:    windowSize,
		HopSize:       hopSize,
	}
}
