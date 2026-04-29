package util

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
	"time"

	gen2brainwebp "github.com/gen2brain/webp"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

type CollageOptions struct {
	TileHeight int
	OutputPath string
}

func DecodeImage(data []byte) (image.Image, error) {
	r := bytes.NewReader(data)

	ct := detectContentType(data)
	switch {
	case strings.HasPrefix(ct, "image/webp"):
		return webp.Decode(r)
	case strings.HasPrefix(ct, "image/jpeg"):
		return jpeg.Decode(r)
	case strings.HasPrefix(ct, "image/png"):
		return png.Decode(r)
	default:
		return nil, fmt.Errorf("unsupported image type: %s", ct)
	}
}

func CreateCollage(tiles []image.Image, opts CollageOptions) (string, error) {
	log := Logger("collage")
	start := time.Now()
	defer func() {
		log.Debug("collage complete", "tiles", len(tiles), "duration", time.Since(start))
	}()

	type sizedTile struct {
		img    image.Image
		width  int
		height int
	}

	resized := make([]sizedTile, 0, len(tiles))
	totalWidth := 0

	for _, tile := range tiles {
		if tile == nil {
			continue
		}

		b := tile.Bounds()
		scale := float64(opts.TileHeight) / float64(b.Dy())
		newW := int(float64(b.Dx())*scale + 0.5)
		newH := opts.TileHeight

		scaled := resizeImage(tile, newW, newH)
		resized = append(resized, sizedTile{img: scaled, width: newW, height: newH})
		totalWidth += newW
	}

	if len(resized) == 0 {
		return "", fmt.Errorf("no tiles to collage")
	}

	canvas := image.NewRGBA(image.Rect(0, 0, totalWidth, opts.TileHeight))

	x := 0
	for _, t := range resized {
		draw.Draw(canvas, image.Rect(x, 0, x+t.width, t.height), t.img, image.Point{}, draw.Over)
		x += t.width
	}

	path := opts.OutputPath + ".webp"
	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("creating collage file: %w", err)
	}
	defer f.Close()

	if err := gen2brainwebp.Encode(f, canvas, gen2brainwebp.Options{Lossless: true, Method: 6}); err != nil {
		return "", fmt.Errorf("encoding webp: %w", err)
	}

	return path, nil
}

func resizeImage(src image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

func detectContentType(data []byte) string {
	if len(data) < 4 {
		return ""
	}

	magic := string(data[:4])
	switch {
	case magic == "RIFF":
		return "image/webp"
	case data[0] == 0xFF && data[1] == 0xD8:
		return "image/jpeg"
	case data[0] == 0x89 && string(data[1:4]) == "PNG":
		return "image/png"
	default:
		return ""
	}
}
