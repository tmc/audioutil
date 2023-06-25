module github.com/tmc/audioutil

go 1.20

require (
	github.com/ggerganov/whisper.cpp/bindings/go v0.0.0-20230606002726-57543c169e27
	github.com/go-audio/audio v1.0.0
	github.com/go-audio/wav v1.1.0
	github.com/gordonklaus/portaudio v0.0.0-20221027163845-7c3b689db3cc
)

require github.com/go-audio/riff v1.0.0 // indirect

replace github.com/ggerganov/whisper.cpp/bindings/go => github.com/tmc/whisper.cpp/bindings/go v0.0.0-20230624233940-156931e468dd
