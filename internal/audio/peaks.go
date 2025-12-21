package audio

import "sort"

const (
	minFreqBin       = 26 // Ignore frequencies below ~250Hz (at 44.1kHz/4096)
	maxPeaksPerFrame = 3  // Only keep strongest peaks per time frame
)

func FindPeaks(spec *Spectrogram, threshold float32, neighborhoodSize int) []Peak {
	var peaks []Peak

	for t := 0; t < len(spec.Data); t++ {
		var framePeaks []Peak

		startF := neighborhoodSize
		if minFreqBin > startF {
			startF = minFreqBin
		}

		for f := startF; f < len(spec.Data[t])-neighborhoodSize; f++ {
			magnitude := spec.Data[t][f]
			if magnitude < threshold {
				continue
			}

			isPeak := true
			for dt := -neighborhoodSize; dt <= neighborhoodSize; dt++ {
				for df := -neighborhoodSize; df <= neighborhoodSize; df++ {
					if dt == 0 && df == 0 {
						continue
					}
					nt, nf := t+dt, f+df
					if nt >= 0 && nt < len(spec.Data) &&
						nf >= 0 && nf < len(spec.Data[nt]) &&
						spec.Data[nt][nf] >= magnitude {
						isPeak = false
						break
					}
				}
				if !isPeak {
					break
				}
			}

			if isPeak {
				framePeaks = append(framePeaks, Peak{
					TimeIndex: t,
					FreqIndex: f,
					Magnitude: magnitude,
				})
			}
		}

		// Keep only strongest peaks for this time frame
		if len(framePeaks) > maxPeaksPerFrame {
			sort.Slice(framePeaks, func(i, j int) bool {
				return framePeaks[i].Magnitude > framePeaks[j].Magnitude
			})
			framePeaks = framePeaks[:maxPeaksPerFrame]
		}

		peaks = append(peaks, framePeaks...)
	}

	return peaks
}
