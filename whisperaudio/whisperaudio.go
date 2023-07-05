package whisperaudio

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/gordonklaus/portaudio"
	"github.com/tmc/audioutil/whisperutil"
)

const (
	channels   = 1
	bufferSize = 2048
)

// WhisperAudio is a wrapper around the whisper library and portaudio.
type WhisperAudio struct {
	model    whisper.Model
	mctx     whisper.Context
	stream   *portaudio.Stream
	inBuffer []float32
}

// New creates a new WhisperAudio instance.
func New(opts ...whisperutil.Option) (*WhisperAudio, error) {
	// Initialize portaudio
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("could not initialize portaudio: %w", err)
	}

	// Initialize whisper model
	modelPath, err := whisperutil.GetModelPath(opts...)
	if err != nil {
		return nil, fmt.Errorf("could not get model path: %w", err)
	}

	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, fmt.Errorf("could not initialize model: %w", err)
	}

	mctx, err := model.NewContext()
	if err != nil {
		return nil, fmt.Errorf("could not initialize context: %w", err)
	}
	//mctx.SetLanguage("en")
	//mctx.SetSpeedup(true)

	// Open audio stream
	in := make([]float32, bufferSize*channels)
	stream, err := portaudio.OpenDefaultStream(channels, 0, whisper.SampleRate, bufferSize, in)
	if err != nil {
		return nil, fmt.Errorf("could not open default stream: %w", err)
	}

	// Create WhisperAudio instance
	return &WhisperAudio{
		model:    model,
		mctx:     mctx,
		stream:   stream,
		inBuffer: in,
	}, nil
}

// DumpDeviceInfo dumps the device info.
func DumpDeviceInfo() {
	portaudio.Initialize()

	fmt.Fprintln(os.Stderr, "default input device:")
	in, err := portaudio.DefaultInputDevice()
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "%+v\n", in)
	fmt.Fprintln(os.Stderr, "default output device:")
	out, err := portaudio.DefaultOutputDevice()
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "%+v\n", out)

	fmt.Fprintln(os.Stderr, "devices:")
	devices, err := portaudio.Devices()
	if err != nil {
		panic(err)
	}
	for _, d := range devices {
		fmt.Println(d)
	}
}

// Start starts the audio stream.
func (wa *WhisperAudio) Start() error {
	var err error
	wa.mctx, err = wa.model.NewContext()
	if err != nil {
		return fmt.Errorf("could not initialize context: %w", err)
	}
	if err := wa.stream.Start(); err != nil {
		return fmt.Errorf("could not start stream: %w", err)
	}
	return nil
}

// CollectAudioData collects audio data for the given duration.
func (wa *WhisperAudio) CollectAudioData(duration time.Duration) ([]float32, error) {
	// TODO: don't truncate to seconds.
	seconds := int(duration.Seconds())
	buf := make([]float32, 0, seconds*(whisper.SampleRate/bufferSize))
	for i := 0; i < seconds*(whisper.SampleRate/bufferSize); i++ {
		err := wa.stream.Read()
		if err != nil {
			return nil, fmt.Errorf("could not read from stream: %w", err)
		}
		buf = append(buf, wa.inBuffer...)
	}
	return buf, nil
}

// Stop stops the audio stream.
func (wa *WhisperAudio) Stop() error {
	if err := wa.stream.Stop(); err != nil {
		return fmt.Errorf("could not stop stream: %w", err)
	}
	return nil
}

// Transcribe transcribes the given audio data.
func (wa *WhisperAudio) Transcribe(buf []float32) (string, error) {
	if err := wa.mctx.Process(buf, nil, func(p int) {
		fmt.Fprintf(os.Stderr, "progress: %d%%\n", p)
	}); err != nil {
		return "", fmt.Errorf("could not process audio: %w", err)
	}
	result := ""
	for {
		s, err := wa.mctx.NextSegment()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("could not get next segment: %w", err)
		}
		result += s.Text
	}
	return result, nil
}
