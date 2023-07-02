package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/objc"
	"github.com/tmc/audioutil/whisperaudio"
)

const (
	NSEventModifierFlagCommand = 1 << 20
	VKControl                  = 0x3B
	VKCommand                  = 0x37
	VKOption                   = 0x3A
)

var (
	defaultTimeout = 30 * time.Second
)

type App struct {
	listeningToggle chan struct{}
	wa              *whisperaudio.WhisperAudio
}

func main() {
	runtime.LockOSThread()
	ctx := context.Background()
	app, err := newApp()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	if err := app.run(ctx); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func newApp() (*App, error) {
	wa, err := whisperaudio.New()
	if err != nil {
		return nil, fmt.Errorf("could not create whisperaudio: %w", err)
	}
	return &App{
		listeningToggle: make(chan struct{}, 1),
		wa:              wa,
	}, nil
}

func (app *App) run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go app.runMainLoop(ctx)
	app.runNSApp(ctx)
	return nil
}

func (app *App) runMainLoop(ctx context.Context) {
	var (
		listening        bool
		listeningTimeout <-chan time.Time
		audioBuffer      []float32
	)
	fmt.Println("ready")
	for {
		select {
		case <-app.listeningToggle:
			listening = !listening
			if listening {
				listeningTimeout = time.After(defaultTimeout)
				fmt.Println("listening...")
				audioBuffer = nil
				err := app.wa.Start()
				if err != nil {
					log.Printf("error starting whisperaudio: %v", err)
				}
			} else {
				fmt.Println("transcribing...")
				if err := app.wa.Stop(); err != nil {
					log.Printf("error stopping whisperaudio: %v", err)
				}
				t1 := time.Now()
				text, err := app.wa.Transcribe(audioBuffer)
				if err != nil {
					log.Printf("error transcribing: %v", err)
					continue
				}
				fmt.Printf("transcribed: %q in %v\n", text, time.Since(t1))
				robotgo.TypeStr(text)
			}
		case <-listeningTimeout:
			if listening {
				app.listeningToggle <- struct{}{}
			}
		case <-ctx.Done():
			fmt.Println("done")
			return
		default:
			if !listening {
				continue
			}
			buf, err := app.wa.CollectAudioData(time.Second)
			if err != nil {
				log.Printf("error collecting audio data: %v", err)
				continue
			}
			audioBuffer = append(audioBuffer, buf...)

		}
	}
}

func (app *App) runNSApp(ctx context.Context) {
	nsApp := cocoa.NSApp_WithDidLaunch(func(n objc.Object) {
		events := make(chan cocoa.NSEvent, 64)
		go app.handleEvents(events)
		cocoa.NSEvent_GlobalMonitorMatchingMask(cocoa.NSEventMaskAny, events)
	})
	nsApp.ActivateIgnoringOtherApps(true)
	nsApp.Run()
}

// handleEvents handles global events
func (app *App) handleEvents(events chan cocoa.NSEvent) {
	for {
		e := <-events
		typ := e.Get("type").Int()
		if typ != cocoa.NSEventTypeFlagsChanged {
			continue
		}
		app.manageListeningState(e)
	}
}

// manageListeningState toggles listening state when the user presses cmd+ctrl
func (app *App) manageListeningState(e cocoa.NSEvent) {
	keyCode := e.Get("keyCode").Int()
	modifierFlags := e.Get("modifierFlags").Int()
	cmdDown := modifierFlags&NSEventModifierFlagCommand != 0
	keyUp := !(modifierFlags&0x1 != 0)
	if (keyCode == VKControl) && cmdDown && keyUp {
		app.listeningToggle <- struct{}{}
	}
}
