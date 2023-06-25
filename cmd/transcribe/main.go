package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/gordonklaus/portaudio"
)

const (
	Channels   = 1
	BufferSize = 2048
)

var (
	FlagDuration = 5 * time.Second
)

func main() {
	flag.Parse()
	err := run()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	if err := initializePortaudio(); err != nil {
		return err
	}
	defer portaudio.Terminate()

	model, err := initializeWhisper()
	if err != nil {
		return fmt.Errorf("could not initialize whisper: %w", err)
	}

	mctx, err := model.NewContext()
	if err != nil {
		return fmt.Errorf("could not initialize context: %w", err)
	}

	stream, in, err := openAudioStream()
	if err != nil {
		return fmt.Errorf("could not open audio stream: %w", err)
	}
	defer stream.Close()

	err = startAudioStream(stream)
	if err != nil {
		return fmt.Errorf("could not start audio stream: %w", err)
	}
	fmt.Fprintln(os.Stderr, "listening.")

	buf, err := collectAudioData(stream, in)
	if err != nil {
		return fmt.Errorf("could not collect audio data: %w", err)
	}

	err = stopAudioStream(stream)
	if err != nil {
		return fmt.Errorf("could not stop audio stream: %w", err)
	}

	fmt.Fprintln(os.Stderr, "transcribing.")
	err = processAudioData(mctx, buf)
	if err != nil {
		return fmt.Errorf("could not process audio data: %w", err)
	}
	fmt.Println()
	return nil
}

func initializePortaudio() error {
	err := portaudio.Initialize()
	if err != nil {
		return fmt.Errorf("could not initialize portaudio: %w", err)
	}
	return nil
}

func initializeWhisper() (whisper.Model, error) {
	modelPath, err := getModelPath()
	if err != nil {
		return nil, err
	}
	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, fmt.Errorf("could not initialize model: %w", err)
	}
	//log.Println("Sample rate:", whisper.SampleRate, whisper.SampleBits)
	return model, nil
}

func openAudioStream() (*portaudio.Stream, []float32, error) {
	in := make([]float32, BufferSize*Channels)
	stream, err := portaudio.OpenDefaultStream(Channels, 0, whisper.SampleRate, BufferSize, in)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open default stream: %w", err)
	}
	return stream, in, nil
}

func startAudioStream(stream *portaudio.Stream) error {
	if err := stream.Start(); err != nil {
		return fmt.Errorf("could not start stream: %w", err)
	}
	return nil
}

func collectAudioData(stream *portaudio.Stream, in []float32) ([]float32, error) {
	seconds := int(FlagDuration.Seconds())
	buf := make([]float32, 0, seconds*(whisper.SampleRate/BufferSize))
	for i := 0; i < seconds*(whisper.SampleRate/BufferSize); i++ {
		err := stream.Read()
		if err != nil {
			return nil, fmt.Errorf("could not read from stream: %w", err)
		}
		buf = append(buf, in...)
	}
	return buf, nil
}

func stopAudioStream(stream *portaudio.Stream) error {
	if err := stream.Stop(); err != nil {
		return fmt.Errorf("could not stop stream: %w", err)
	}
	return nil
}

func processAudioData(mctx whisper.Context, buf []float32) error {
	if err := mctx.Process(buf, nil); err != nil {
		return fmt.Errorf("could not process audio: %w", err)
	}

	for {
		s, err := mctx.NextSegment()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("could not get next segment: %w", err)
		}
		fmt.Print(s.Text)
	}

	return nil
}

func getModelPath() (string, error) {
	cd, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("could not get user cache directory: %w", err)
	}
	path := filepath.Join(cd, "whisper.cpp", "ggml-base.bin")
	return path, nil
}
