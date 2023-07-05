package wavutil

import (
	"fmt"
	"io"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/transforms"
	"github.com/go-audio/wav"
)

const (
	defaultWavBitDepth = 24
)

// SaveWAV saves the given data as a WAV file with the given sample rate.
func SaveWAV(filename string, data []float32, sampleRate int) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer f.Close()

	if err := WriteWAV(f, data, sampleRate); err != nil {
		return fmt.Errorf("could not write wav file: %w", err)
	}

	return nil
}

// WriteWAV writes the given data as a WAV file with the given sample rate to the given io.WriteSeeker.
func WriteWAV(o io.WriteSeeker, data []float32, sampleRate int) error {
	// write header
	const (
		audioFormatPCM = 1
		channels       = 1
	)
	buf := &audio.Float32Buffer{
		Format: &audio.Format{
			NumChannels: channels,
			SampleRate:  sampleRate,
		},
		Data: data,
	}
	transforms.PCMScaleF32(buf, defaultWavBitDepth)
	e := wav.NewEncoder(o, sampleRate, defaultWavBitDepth, channels, audioFormatPCM)
	if err := e.Write(buf.AsIntBuffer()); err != nil {
		return err
	}
	return e.Close()
}
