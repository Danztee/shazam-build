package audio

type AudioData struct {
	Samples    []int16
	SampleRate int
	Channels   int
	Duration   float64
}

type Spectrogram struct {
	Data          [][]float32
	TimeBins      []float64
	FrequencyBins []float64
	SampleRate    int
	WindowSize    int
	HopSize       int
}

type Peak struct {
	TimeIndex int
	FreqIndex int
	Magnitude float32
}

type Fingerprint struct {
	Hash         int64
	TimeOffsetMs int
}

type IndexedFingerprint struct {
	Hash         int64
	SongID       int
	TimeOffsetMs int
}
