package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	webp "golang.org/x/image/webp"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	imgURL := `https://songstitch.art/collage?` +
		`username=grumpylama&method=album&period=7day&artist=false` +
		`&album=false&playcount=false&rows=1&columns=5&fontsize=15` +
		`&textlocation=bottomcentre&webp=false`

	imageData, mimeType, err := fetchImage(ctx, imgURL)
	if err != nil {
		log.Fatalf("Failed to fetch image: %v", err)
	}

	base64Data, err := encodeImageToBase64(imageData, mimeType)
	if err != nil {
		log.Fatalf("Failed to encode image: %v", err)
	}

	markdown, err := generateMarkdown(base64Data, mimeType)
	if err != nil {
		log.Fatalf("Failed to generate markdown: %v", err)
	}

	// Write the generated README to file
	err = os.WriteFile("README.md", []byte(markdown), 0644)
	if err != nil {
		log.Fatalf("Failed to write README.md: %v", err)
	}

	fmt.Println("README.md updated successfully")
}

func fetchImage(ctx context.Context, url string) ([]byte, string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading response body: %w", err)
	}

	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}

	return data, mimeType, nil
}

func encodeImageToBase64(data []byte, mimeType string) (string, error) {
	// First, validate the image data integrity by attempting to decode it
	// This ensures the data is valid before we use it
	var decodeErr error

	switch {
	case strings.HasPrefix(mimeType, "image/webp"):
		_, decodeErr = webp.Decode(bytes.NewReader(data))
	case strings.HasPrefix(mimeType, "image/jpeg"):
		_, decodeErr = jpeg.Decode(bytes.NewReader(data))
	case strings.HasPrefix(mimeType, "image/png"):
		_, decodeErr = png.Decode(bytes.NewReader(data))
	default:
		return "", fmt.Errorf("unsupported image type: %s", mimeType)
	}

	if decodeErr != nil {
		return "", fmt.Errorf("invalid image data for %s: %w", mimeType, decodeErr)
	}

	// Image data is valid, now preserve original quality by encoding
	// the original bytes directly to base64 (no reencoding)
	return base64.StdEncoding.EncodeToString(data), nil
}

func generateMarkdown(base64Data, mimeType string) (string, error) {
	readmeContent, err := os.ReadFile("tpl/README.md.tmpl")
	if err != nil {
		return "", fmt.Errorf("reading README.md.tmpl: %w", err)
	}

	// Create the embedded image data URI
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	// Parse and execute the template
	tmpl, err := template.New("readme").Parse(string(readmeContent))
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	data := struct {
		ImageURL       string
		BuildTimestamp string
	}{
		ImageURL:       dataURI,
		BuildTimestamp: time.Now().Format(time.RFC3339),
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}
