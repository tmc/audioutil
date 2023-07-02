module github.com/tmc/audioutil

go 1.20

require (
	github.com/ggerganov/whisper.cpp/bindings/go v0.0.0-20230606002726-57543c169e27
	github.com/go-audio/audio v1.0.0
	github.com/go-audio/wav v1.1.0
	github.com/go-vgo/robotgo v0.100.10
	github.com/gordonklaus/portaudio v0.0.0-20221027163845-7c3b689db3cc
	github.com/progrium/macdriver v0.3.0
	golang.design/x/hotkey v0.4.1
)

require (
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/go-audio/riff v1.0.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/lxn/win v0.0.0-20210218163916-a377121e959e // indirect
	github.com/otiai10/gosseract v2.2.1+incompatible // indirect
	github.com/robotn/gohook v0.31.3 // indirect
	github.com/robotn/xgb v0.0.0-20190912153532-2cb92d044934 // indirect
	github.com/robotn/xgbutil v0.0.0-20190912154524-c861d6f87770 // indirect
	github.com/shirou/gopsutil v3.21.10+incompatible // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	github.com/vcaesar/gops v0.21.3 // indirect
	github.com/vcaesar/imgo v0.30.0 // indirect
	github.com/vcaesar/keycode v0.10.0 // indirect
	github.com/vcaesar/tt v0.20.0 // indirect
	golang.org/x/image v0.0.0-20211028202545-6944b10bf410 // indirect
	golang.org/x/sys v0.0.0-20211123173158-ef496fb156ab // indirect
)

replace github.com/ggerganov/whisper.cpp/bindings/go => github.com/tmc/whisper.cpp/bindings/go v0.0.0-20230624233940-156931e468dd
