package util

import (
	"encoding/json"
	"fmt"
	"os"
)

const stateFile = "data/state.json"

type State struct {
	LastfmImagePath string `json:"lastfm_image_path"`
	ImdbImagePath   string `json:"imdb_image_path"`
}

func LoadState() (State, error) {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, nil
		}
		return State{}, fmt.Errorf("reading state: %w", err)
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, fmt.Errorf("parsing state: %w", err)
	}

	return s, nil
}

func SaveState(s State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("writing state: %w", err)
	}

	return nil
}
