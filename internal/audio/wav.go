package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func ReadWAV(filePath string) (*AudioData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var riffHeader [12]byte
	if _, err := io.ReadFull(file, riffHeader[:]); err != nil {
		return nil, fmt.Errorf("failed to read RIFF header: %w", err)
	}

	if string(riffHeader[0:4]) != "RIFF" || string(riffHeader[8:12]) != "WAVE" {
		return nil, fmt.Errorf("invalid WAV file")
	}

	var fmtChunkID [4]byte
	if _, err := io.ReadFull(file, fmtChunkID[:]); err != nil {
		return nil, fmt.Errorf("failed to read fmt chunk ID: %w", err)
	}

	if string(fmtChunkID[:]) != "fmt " {
		return nil, fmt.Errorf("fmt chunk not found")
	}

	var fmtChunkSize [4]byte
	if _, err := io.ReadFull(file, fmtChunkSize[:]); err != nil {
		return nil, fmt.Errorf("failed to read fmt chunk size: %w", err)
	}

	fmtSize := binary.LittleEndian.Uint32(fmtChunkSize[:])
	if fmtSize < 16 {
		return nil, fmt.Errorf("invalid fmt chunk size")
	}

	fmtData := make([]byte, fmtSize)
	if _, err := io.ReadFull(file, fmtData[:]); err != nil {
		return nil, fmt.Errorf("failed to read fmt data: %w", err)
	}

	audioFormat := binary.LittleEndian.Uint16(fmtData[0:2])
	if audioFormat != 1 {
		return nil, fmt.Errorf("only PCM format supported")
	}

	channels := int(binary.LittleEndian.Uint16(fmtData[2:4]))
	sampleRate := int(binary.LittleEndian.Uint32(fmtData[4:8]))
	bitsPerSample := int(binary.LittleEndian.Uint16(fmtData[14:16]))

	var dataSize uint32
	found := false
	for i := 0; i < 20; i++ {
		var chunkID [4]byte
		if _, err := io.ReadFull(file, chunkID[:]); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return nil, fmt.Errorf("failed to read chunk ID: %w", err)
		}

		var chunkSize [4]byte
		if _, err := io.ReadFull(file, chunkSize[:]); err != nil {
			return nil, fmt.Errorf("failed to read chunk size: %w", err)
		}

		size := binary.LittleEndian.Uint32(chunkSize[:])

		if string(chunkID[:]) == "data" {
			dataSize = size
			found = true
			break
		}

		if _, err := file.Seek(int64(size), io.SeekCurrent); err != nil {
			return nil, fmt.Errorf("failed to skip chunk: %w", err)
		}
	}

	if !found || dataSize == 0 {
		return nil, fmt.Errorf("data chunk not found")
	}

	if bitsPerSample != 16 {
		return nil, fmt.Errorf("only 16-bit samples supported")
	}

	currentPos, _ := file.Seek(0, io.SeekCurrent)
	fileInfo, _ := file.Stat()
	availableBytes := uint32(fileInfo.Size() - currentPos)
	if availableBytes > dataSize {
		availableBytes = dataSize
	}
	if availableBytes == 0 {
		return nil, fmt.Errorf("no audio data available")
	}

	bytesPerSample := 2 * channels
	numSamples := int(availableBytes) / bytesPerSample
	if numSamples == 0 {
		return nil, fmt.Errorf("insufficient data")
	}

	pcmData := make([]byte, numSamples*bytesPerSample)
	bytesRead, err := io.ReadFull(file, pcmData)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read PCM data: %w", err)
	}

	if bytesRead < len(pcmData) {
		pcmData = pcmData[:bytesRead]
		numSamples = bytesRead / bytesPerSample
		if numSamples == 0 {
			return nil, fmt.Errorf("insufficient data read")
		}
	}

	samples := make([]int16, numSamples*channels)
	for i := 0; i < numSamples*channels; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(pcmData[i*2 : (i+1)*2]))
	}

	if channels == 2 {
		mono := make([]int16, numSamples)
		for i := 0; i < numSamples; i++ {
			mono[i] = (samples[i*2] + samples[i*2+1]) / 2
		}
		samples = mono
		channels = 1
	}

	return &AudioData{
		Samples:    samples,
		SampleRate: sampleRate,
		Channels:   channels,
		Duration:   float64(numSamples) / float64(sampleRate),
	}, nil
}
