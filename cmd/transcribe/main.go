// Command transcribe records audio from the microphone and transcribes it.
//
// Usage of transcribe:
//
//	-duration duration
//	  	duration of audio to transcribe (default 5s)
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/tmc/audioutil/whisperaudio"
)

var (
	flagDuration = flag.Duration("duration", 5*time.Second, "duration of audio to transcribe")
)

func main() {
	flag.Parse()
	err := run()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	duration := *flagDuration
	wa, err := whisperaudio.New()
	if err != nil {
		return fmt.Errorf("could not initialize whisperaudio: %w", err)
	}
	defer wa.Stop()

	if err = wa.Start(); err != nil {
		return fmt.Errorf("could not start whisperaudio: %w", err)
	}

	data, err := wa.CollectAudioData(duration)
	if err != nil {
		return fmt.Errorf("could not collect audio data: %w", err)
	}

	text, err := wa.Transcribe(data)
	if err != nil {
		return fmt.Errorf("could not transcribe audio data: %w", err)
	}
	fmt.Println(text)
	return nil
}
