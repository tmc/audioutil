package whisperutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultModelName = "ggml-base.bin"
	CacheDirName     = "whisper.cpp"
)

// ModelPathOptions is used to configure the model path.
type ModelPathOptions struct {
	ModelName string
	AutoFetch bool
}

// Option is a function that configures a ModelPathOptions.
type Option func(*ModelPathOptions)

// WithModelName sets the model name to use.
func WithModelName(modelName string) Option {
	return func(mpo *ModelPathOptions) {
		mpo.ModelName = modelName
		// add prefix and suffix if not present
		if !strings.HasPrefix(mpo.ModelName, "ggml-") {
			mpo.ModelName = "ggml-" + mpo.ModelName
		}
		if filepath.Ext(mpo.ModelName) != ".bin" {
			mpo.ModelName += ".bin"
		}
	}
}

// WithAutoFetch enables auto-fetching of the model if it is not found.
func WithAutoFetch() Option {
	return func(mpo *ModelPathOptions) {
		mpo.AutoFetch = true
	}
}

// GetModelPath returns the path to the model file.
func GetModelPath(opts ...Option) (string, error) {
	options := ModelPathOptions{
		ModelName: DefaultModelName, // Default model name
		AutoFetch: false,            // Default AutoFetch
	}
	for _, opt := range opts {
		opt(&options)
	}

	cd, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("could not get user cache directory: %w", err)
	}
	path := filepath.Join(cd, CacheDirName, options.ModelName)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if options.AutoFetch {
			fmt.Fprintln(os.Stderr, "Model not found, trying to fetch it...")
			// not implemented ddy et
			return path, autoFetch(path, options.ModelName)
		}
	}
	return path, nil
}
