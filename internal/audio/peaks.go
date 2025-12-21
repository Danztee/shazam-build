package audio

func FindPeaks(spec *Spectrogram, threshold float32, neighborhoodSize int) []Peak {
	var peaks []Peak

	for t := 0; t < len(spec.Data); t++ {
		for f := neighborhoodSize; f < len(spec.Data[t])-neighborhoodSize; f++ {
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
				peaks = append(peaks, Peak{
					TimeIndex: t,
					FreqIndex: f,
					Magnitude: magnitude,
				})
			}
		}
	}

	return peaks
}
