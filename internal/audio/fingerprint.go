package audio

func GenerateFingerprints(peaks []Peak, spec *Spectrogram, timeRange int) []Fingerprint {
	var fingerprints []Fingerprint
	const maxPairsPerAnchor = 5

	for i, anchor := range peaks {
		pairCount := 0
		for j := i + 1; j < len(peaks) && pairCount < maxPairsPerAnchor; j++ {
			target := peaks[j]
			timeDiff := target.TimeIndex - anchor.TimeIndex
			if timeDiff > timeRange {
				break
			}
			if timeDiff <= 0 {
				continue
			}

			hash := int64(anchor.FreqIndex&0x3FF)<<22 |
				int64(target.FreqIndex&0x3FF)<<12 |
				int64(timeDiff&0xFFF)

			anchorTimeMs := int(float64(anchor.TimeIndex) *
				float64(spec.HopSize) / float64(spec.SampleRate) * 1000)

			fingerprints = append(fingerprints, Fingerprint{
				Hash:         hash,
				AnchorTimeMs: anchorTimeMs,
			})
			pairCount++
		}
	}

	return fingerprints
}
