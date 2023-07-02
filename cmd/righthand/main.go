package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/objc"
	"github.com/tmc/audioutil/whisperaudio"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

const (
	// NSEventModifierFlagCommand is the command key modifier flag.
	NSEventModifierFlagCommand = 1 << 20
	// VKControl is the virtual key code for the control key.
	VKControl = 0x3B
	// VKCommand is the virtual key code for the command key.
	VKCommand = 0x37
	// VKOption is the virtual key code for the option key.
	VKOption = 0x3A
)

var (
	// defaultTimeout is the default timeout for listening.
	defaultTimeout = 30 * time.Second
)

// App is the main application.
type App struct {
	listeningToggle chan struct{}
	wa              *whisperaudio.WhisperAudio
	llm             llms.ChatLLM
	cfg             *RightHandConfig
}

// main is the entrypoint.
func main() {
	runtime.LockOSThread()
	flag.Parse()
	ctx := context.Background()
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
	}
	app, err := newApp(cfg)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	if err := app.run(ctx); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// newApp creates a new app.
func newApp(cfg RightHandConfig) (*App, error) {
	wa, err := whisperaudio.New()
	if err != nil {
		return nil, fmt.Errorf("could not create whisperaudio: %w", err)
	}
	cllm, err := openai.NewChat()
	if err != nil {
		return nil, fmt.Errorf("could not create chat LLM: %w", err)
	}
	return &App{
		listeningToggle: make(chan struct{}, 1),
		wa:              wa,
		llm:             cllm,
		cfg:             &cfg,
	}, nil
}

// run runs the app.
func (app *App) run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go app.runMainLoop(ctx)
	app.runNSApp(ctx)
	return nil
}

// runMainLoop runs the main loop.
func (app *App) runMainLoop(ctx context.Context) {
	var (
		listening        bool
		listeningTimeout <-chan time.Time
		audioBuffer      []float32
	)
	fmt.Println("righthand: ready")
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
				if text != "" {
					go app.handleText(ctx, text)
				}
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

// runNSApp runs the NSApp.
func (app *App) runNSApp(ctx context.Context) {
	nsApp := cocoa.NSApp_WithDidLaunch(func(n objc.Object) {
		events := make(chan cocoa.NSEvent, 64)
		go app.handleEvents(events)
		cocoa.NSEvent_GlobalMonitorMatchingMask(cocoa.NSEventMaskAny, events)
	})
	nsApp.ActivateIgnoringOtherApps(true)
	nsApp.Run()
}

// handleEvents handles global events.
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

// manageListeningState toggles listening state.
func (app *App) manageListeningState(e cocoa.NSEvent) {
	keyCode := e.Get("keyCode").Int()
	modifierFlags := e.Get("modifierFlags").Int()
	cmdDown := modifierFlags&NSEventModifierFlagCommand != 0
	keyUp := !(modifierFlags&0x1 != 0)
	if (keyCode == VKControl) && cmdDown && keyUp {
		app.listeningToggle <- struct{}{}
	}
}

var systemPrompt = `You are an AI assistant that interprets spoken language 
and translates it into commands or text inputs for various applications. 

Your current active program is %v. Adjust your interpretation based on this context.

When interpreting commands, please indicate modifier keys such as Command, Option, Shift, 
or Control using curly braces. For instance, use '{Command}+t' for opening a new tab.

Only print the modified text or print the original input if you are unsure. 

Do not try to interpret or answer the prompt, but merely contextualize it for the active application.`

// handleText handles text.
func (app *App) handleText(ctx context.Context, text string) {
	activeApp := fmt.Sprint(cocoa.NSWorkspace_sharedWorkspace().FrontmostApplication().LocalizedName())
	fmt.Println("active app:", activeApp)

	messages := []schema.ChatMessage{
		schema.SystemChatMessage{
			Text: fmt.Sprintf(systemPrompt, activeApp),
		},
	}

	// check for few-shot examples for the active app from the config:
	// TODO(tmc): this would be faster as a map
	nExamples := 0
	for _, prog := range app.cfg.Programs {
		if prog.Program != activeApp {
			continue
		}
		for _, example := range prog.Examples {
			messages = append(messages, schema.HumanChatMessage{Text: example.Input})
			messages = append(messages, schema.AIChatMessage{Text: example.Output})
		}
		nExamples = len(prog.Examples)
	}

	fmt.Fprintf(os.Stderr, "righthand: using %v few-shot examples for %v\n", nExamples, activeApp)

	// append the human message:
	messages = append(messages, schema.HumanChatMessage{Text: text})

	llmText, err := app.llm.Call(ctx, messages)
	if err != nil {
		log.Printf("error calling LLM: %v", err)
		return
	}
	fmt.Println("response:", llmText)
	robotgo.TypeStr(llmText)
}
