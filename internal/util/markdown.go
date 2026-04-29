package util

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
	"time"
)

type templateData struct {
	LastfmImagePath string
	ImdbImagePath   string
	BuildTimestamp   string
}

func generateMarkdown(tmplPath string, data templateData) (string, error) {
	content, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("reading template: %w", err)
	}

	tmpl, err := template.New("readme").Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	data.BuildTimestamp = time.Now().Format(time.RFC3339)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

func writeFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func WriteReadme(state State) error {
	log := Logger("markdown")
	start := time.Now()
	defer func() {
		log.Debug("readme generated", "duration", time.Since(start))
	}()

	markdown, err := generateMarkdown("tpl/README.md.tmpl", templateData{
		LastfmImagePath: state.LastfmImagePath,
		ImdbImagePath:   state.ImdbImagePath,
	})
	if err != nil {
		return fmt.Errorf("generating markdown: %w", err)
	}

	return writeFile("README.md", markdown)
}
