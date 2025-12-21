package audio

func GenerateFingerprints(peaks []Peak, spec *Spectrogram, timeRange int) []Fingerprint {
	var fingerprints []Fingerprint
	const maxPairsPerAnchor = 10

	for i, anchor := range peaks {
		if anchor.FreqIndex >= 1024 {
			continue
		}
		pairCount := 0
		for j := i + 1; j < len(peaks) && pairCount < maxPairsPerAnchor; j++ {
			target := peaks[j]
			if target.FreqIndex >= 1024 {
				continue
			}
			timeDiff := target.TimeIndex - anchor.TimeIndex
			if timeDiff > timeRange {
				break
			}
			if timeDiff <= 0 {
				continue
			}

			// Pack into 32-bit hash:
			// 10 bits anchor freq | 10 bits target freq | 12 bits time delta
			hash := int64(anchor.FreqIndex&0x3FF)<<22 |
				int64(target.FreqIndex&0x3FF)<<12 |
				int64(timeDiff&0xFFF)

			anchorTimeMs := int(float64(anchor.TimeIndex) *
				float64(spec.HopSize) / float64(spec.SampleRate) * 1000)

			fingerprints = append(fingerprints, Fingerprint{
				Hash:         hash,
				TimeOffsetMs: anchorTimeMs,
			})
			pairCount++
		}
	}

	return fingerprints
}
