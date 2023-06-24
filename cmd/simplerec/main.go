package main

import (
	"fmt"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate = 44100
	channels   = 1
	seconds    = 3
	bufferSize = 512
)

func printAudioDevices() error {
	devices, err := portaudio.Devices()
	if err != nil {
		return fmt.Errorf("Error: Could not get devices - %v", err)
	}

	for i, device := range devices {
		fmt.Printf("Device %d: %s\n", i, device.Name)
		fmt.Printf("  MaxInputChannels: %d\n", device.MaxInputChannels)
		fmt.Printf("  MaxOutputChannels: %d\n", device.MaxOutputChannels)
		fmt.Printf("  DefaultLowInputLatency: %.2f ms\n", device.DefaultLowInputLatency.Seconds()*1000)
		fmt.Printf("  DefaultLowOutputLatency: %.2f ms\n", device.DefaultLowOutputLatency.Seconds()*1000)
		fmt.Printf("  DefaultHighInputLatency: %.2f ms\n", device.DefaultHighInputLatency.Seconds()*1000)
		fmt.Printf("  DefaultHighOutputLatency: %.2f ms\n", device.DefaultHighOutputLatency.Seconds()*1000)
		fmt.Printf("  DefaultSampleRate: %.2f Hz\n", device.DefaultSampleRate)
	}

	return nil
}

func main() {
	err := portaudio.Initialize()
	if err != nil {
		fmt.Printf("Error: Could not initialize portaudio - %v", err)
		return
	}
	defer portaudio.Terminate()
	printAudioDevices()

	in := make([]int32, bufferSize*channels)
	stream, err := portaudio.OpenDefaultStream(channels, 0, sampleRate, bufferSize, in)
	if err != nil {
		fmt.Printf("Error: Could not open default stream - %v", err)
		return
	}
	defer stream.Close()

	outFile, err := os.Create("output.wav")
	if err != nil {
		fmt.Printf("Error: Could not create output file - %v", err)
		return
	}
	defer outFile.Close()

	enc := wav.NewEncoder(outFile, sampleRate, 16, channels, 1)
	if err != nil {
		fmt.Printf("Error: Could not create encoder - %v", err)
		return
	}

	fmt.Println("Recording. Please speak into the microphone.")
	err = stream.Start()
	if err != nil {
		fmt.Printf("Error: Could not start stream - %v", err)
		return
	}

	for i := 0; i < seconds*sampleRate/bufferSize; i++ {
		err = stream.Read()
		if err != nil {
			fmt.Printf("Error: Could not read from stream - %v", err)
			return
		}

		ints := make([]int, len(in))
		for index, val := range in {
			ints[index] = int(val >> 16)
		}
		err = enc.Write(&audio.IntBuffer{Data: ints, Format: &audio.Format{SampleRate: sampleRate, NumChannels: channels}})
		if err != nil {
			fmt.Printf("Error: Could not write to buffer - %v", err)
			return
		}
	}
	stream.Stop()

	err = enc.Close()
	if err != nil {
		fmt.Printf("Error: Could not close encoder - %v", err)
		return
	}

	fmt.Println("Recording done. output.wav saved.")
}
