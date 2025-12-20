package main

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
		`&textlocation=bottomcentre&webp=true`

	imageFilename, err := fetchAndSaveImage(ctx, imgURL)
	if err != nil {
		log.Fatalf("Failed to fetch and save image: %v", err)
	}

	markdown, err := generateMarkdown(imageFilename)
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

func fetchAndSaveImage(ctx context.Context, url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

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

	// Validate image data
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

	// Determine file extension
	var ext string
	switch {
	case strings.HasPrefix(mimeType, "image/webp"):
		ext = ".webp"
	case strings.HasPrefix(mimeType, "image/jpeg"):
		ext = ".jpg"
	case strings.HasPrefix(mimeType, "image/png"):
		ext = ".png"
	default:
		return "", fmt.Errorf("unsupported image type: %s", mimeType)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("lastfm-top-albums%s", ext)
	filepath := filepath.Join("assets/img", filename)

	// Save image to file
	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return "", fmt.Errorf("saving image file: %w", err)
	}

	return filepath, nil
}

func generateMarkdown(imageFilename string) (string, error) {
	readmeContent, err := os.ReadFile("tpl/README.md.tmpl")
	if err != nil {
		return "", fmt.Errorf("reading README.md.tmpl: %w", err)
	}

	// Parse and execute the template
	tmpl, err := template.New("readme").Parse(string(readmeContent))
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	data := struct {
		ImagePath      string
		BuildTimestamp string
	}{
		ImagePath:      imageFilename,
		BuildTimestamp: time.Now().Format(time.RFC3339),
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}
