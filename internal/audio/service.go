package audio

import (
	"context"
	"fmt"
	"log/slog"
)

type Service interface {
	ProcessAudio(ctx context.Context, wavPath string) ([]Fingerprint, error)
}

type svc struct {
	logger *slog.Logger
}

func NewService(logger *slog.Logger) Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &svc{logger: logger}
}

func (s *svc) ProcessAudio(ctx context.Context, wavPath string) ([]Fingerprint, error) {
	audio, err := ReadWAV(wavPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read WAV: %w", err)
	}

	spec := ComputeSpectrogram(audio, DefaultWindowSize, DefaultHopSize)
	peaks := FindPeaks(spec, -30.0, 2)
	fingerprints := GenerateFingerprints(peaks, spec, 50)

	s.logger.Info("processed audio",
		"peaks", len(peaks),
		"fingerprints", len(fingerprints),
		"duration_seconds", audio.Duration)

	return fingerprints, nil
}
