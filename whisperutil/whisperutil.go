package whisperutil

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultModelName = "ggml-base.bin"
	CacheDirName     = "whisper.cpp"
)

type ModelPathOptions struct {
	ModelName string
	AutoFetch bool
}

type Option func(*ModelPathOptions)

func WithModelName(modelName string) Option {
	return func(mpo *ModelPathOptions) {
		mpo.ModelName = modelName
	}
}

func WithAutoFetch(autoFetch bool) Option {
	return func(mpo *ModelPathOptions) {
		mpo.AutoFetch = autoFetch
	}
}

func GetModelPath(opts ...Option) (string, error) {
	options := ModelPathOptions{
		ModelName: DefaultModelName, // Default model name
		AutoFetch: false,            // Default AutoFetch
	}
	for _, opt := range opts {
		opt(&options)
	}

	// Handle AutoFetch option
	if options.AutoFetch {
		// not implemented yet
		return "", fmt.Errorf("AutoFetch is not implemented yet")
	}

	cd, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("could not get user cache directory: %w", err)
	}
	path := filepath.Join(cd, CacheDirName, options.ModelName)
	return path, nil
}
