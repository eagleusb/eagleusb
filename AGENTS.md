# AGENTS.md

## Overview

This is a Go repository containing a CLI tool to gather statistics and write
markdown files with templates.

## Command

```bash
# Format:
  go fmt ./...

# Fix:
  go fix ./...

# Static analysis:
  CGO_ENABLED=0 go vet ./...

# Build (host):
  CGO_ENABLED=0 go build -o bin/stats .

# Build (arm64):
  CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/stats-arm64 .
```

## Dependencies

```bash
# Install/update dependencies
go mod download
go mod tidy
```

## Error Handling

- Uses error wrapping pattern: `fmt.Errorf("context: %w", err)`
- Errors are logged with `log.Fatalf()` for fatal errors
- Errors are printed with `fmt.Printf()` for non-fatal errors

## Struct Tags

- JSON unmarshaling uses `json:"field_name"` tags
- Case-sensitive field matching for API responses

## Image Handling

- Supports WebP, JPEG, and PNG formats
- Uses `golang.org/x/image/webp` for WebP decoding

## CSV Parsing

- Uses standard `encoding/csv` package
- Column indices are hard-coded (e.g., `record[1]` for "Your Rating")
- String trimming for quoted fields: `strings.Trim(record[3], "\"")`

## Naming Conventions

- **Functions**: `camelCase` - public functions start with capital letter
  (`GetLatestRatings`)
- **Variables**: `camelCase`
- **Constants**: Not heavily used; when present, would be `UPPER_SNAKE_CASE`
- **Structs/Types**: `PascalCase` (`Rating`, `TMDBSearchResponse`)
- **Files**: `snake_case.go`
