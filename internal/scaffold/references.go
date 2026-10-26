package scaffold

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const maxReferenceBytes = 8 << 20

type referenceSource struct {
	FileName string
	URL      string
}

var referenceSources = []referenceSource{
	{FileName: "nuxt-llms-full.txt", URL: "https://nuxt.com/llms-full.txt"},
	{FileName: "nuxt-ui-llms-full.txt", URL: "https://ui.nuxt.com/llms-full.txt"},
}

type Downloader func(ctx context.Context, url string) ([]byte, error)

func downloadReference(ctx context.Context, url string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build reference request: %w", err)
	}
	client := &http.Client{Timeout: 90 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: HTTP %d", url, response.StatusCode)
	}
	if response.ContentLength > maxReferenceBytes {
		return nil, fmt.Errorf("download %s: response exceeds %d bytes", url, maxReferenceBytes)
	}

	contents, err := io.ReadAll(io.LimitReader(response.Body, maxReferenceBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", url, err)
	}
	if len(contents) == 0 || len(contents) > maxReferenceBytes {
		return nil, fmt.Errorf("download %s: invalid response size %d", url, len(contents))
	}
	return contents, nil
}
