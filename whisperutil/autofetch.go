package whisperutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	srcUrl  = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main" // The location of the models
	srcExt  = ".bin"
	bufSize = 1 << 20 // 1 MB
)

// urlForModel returns the URL for the given model on huggingface.co
func urlForModel(model string) (string, error) {
	if filepath.Ext(model) != srcExt {
		model += srcExt
	}
	url, err := url.Parse(srcUrl)
	if err != nil {
		return "", err
	} else {
		url.Path = filepath.Join(url.Path, model)
	}
	return url.String(), nil
}

func autoFetch(path, modelName string) error {
	ctx := context.Background()
	u, err := urlForModel(modelName)
	if err != nil {
		return err
	}
	_, err = download(ctx, os.Stderr, u, path)
	return err
}

// download downloads the model from the given URL to the given output directory
func download(ctx context.Context, p io.Writer, model, path string) (string, error) {
	// Create HTTP client
	client := http.Client{
		Timeout: 15 * time.Minute,
	}

	// Initiate the download
	req, err := http.NewRequest("GET", model, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: %s", model, resp.Status)
	}

	// Create output directory, if needed
	dir := filepath.Dir(path)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}

	// If output file exists and is the same size as the model, skip
	if info, err := os.Stat(path); err == nil && info.Size() == resp.ContentLength {
		fmt.Fprintln(p, "Skipping", model, "as it already exists")
		return "", nil
	}

	// Create file
	w, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer w.Close()

	// Report
	fmt.Fprintln(p, "Downloading", model, "to", path)

	// Progressively download the model
	data := make([]byte, bufSize)
	count, pct := int64(0), int64(0)
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ctx.Done():
			// Cancelled, return error
			return path, ctx.Err()
		case <-ticker.C:
			pct = downloadReport(p, pct, count, resp.ContentLength)
		default:
			// Read body
			n, err := resp.Body.Read(data)
			if err != nil {
				downloadReport(p, pct, count, resp.ContentLength)
				if err == io.EOF {
					return path, nil
				}
				return path, err
			} else if m, err := w.Write(data[:n]); err != nil {
				return path, fmt.Errorf("failed to write to %s: %w", path, err)
			} else {
				count += int64(m)
			}
		}
	}
}

func downloadReport(w io.Writer, pct, count, total int64) int64 {
	pct_ := count * 100 / total
	if pct_ > pct {
		fmt.Fprintf(w, "  ...%d MB written (%d%%)\n", count/1e6, pct_)
	}
	return pct_
}
