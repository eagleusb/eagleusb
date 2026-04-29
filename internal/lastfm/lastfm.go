package lastfm

import (
	"context"
	"fmt"
	"time"

	"eagleusb/eagleusb/internal/util"
)

const (
	imgURL = `https://songstitch.art/collage?` +
		`username=grumpylama&method=album&period=7day&artist=false` +
		`&album=false&playcount=false&rows=1&columns=5&fontsize=15` +
		`&textlocation=bottomcentre&webp=true`
)

func Run(verbose bool) error {
	log := util.Logger("lastfm")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Debug("fetching image", "url", imgURL)

	imagePath, err := util.FetchAndSaveImage(ctx, imgURL, "lastfm-top-albums")
	if err != nil {
		return fmt.Errorf("fetching lastfm image: %w", err)
	}

	log.Info("image saved", "path", imagePath)

	state, err := util.LoadState()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	state.LastfmImagePath = imagePath

	if err := util.SaveState(state); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	if err := util.WriteReadme(state); err != nil {
		return err
	}

	log.Info("readme updated")
	return nil
}
