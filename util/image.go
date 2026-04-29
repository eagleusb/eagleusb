package util

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	webp "golang.org/x/image/webp"
)

func FetchAndSaveImage(ctx context.Context, imageURL string, filename string) (string, error) {
	client := NewClient()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	setUserAgent(req)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}

	if err := validateImage(data, mimeType); err != nil {
		return "", err
	}

	ext, err := extFromMIME(mimeType)
	if err != nil {
		return "", err
	}

	fp := filepath.Join("assets/img", filename+ext)

	if err := os.WriteFile(fp, data, 0644); err != nil {
		return "", fmt.Errorf("saving image file: %w", err)
	}

	return fp, nil
}

func validateImage(data []byte, mimeType string) error {
	r := bytes.NewReader(data)
	switch {
	case strings.HasPrefix(mimeType, "image/webp"):
		_, err := webp.Decode(r)
		return err
	case strings.HasPrefix(mimeType, "image/jpeg"):
		_, err := jpeg.Decode(r)
		return err
	case strings.HasPrefix(mimeType, "image/png"):
		_, err := png.Decode(r)
		return err
	default:
		return fmt.Errorf("unsupported image type: %s", mimeType)
	}
}

func extFromMIME(mimeType string) (string, error) {
	switch {
	case strings.HasPrefix(mimeType, "image/webp"):
		return ".webp", nil
	case strings.HasPrefix(mimeType, "image/jpeg"):
		return ".jpg", nil
	case strings.HasPrefix(mimeType, "image/png"):
		return ".png", nil
	default:
		return "", fmt.Errorf("unsupported image type: %s", mimeType)
	}
}
