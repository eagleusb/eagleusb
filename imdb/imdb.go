package imdb

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"eagleusb/eagleusb/util"
)

const (
	tmdbTokenEnv  = "TMDB_ACCESS_TOKEN"
	tmdbFindURL   = "https://api.themoviedb.org/3/find/%s?external_source=imdb_id"
	tmdbImageBase = "https://image.tmdb.org/t/p/w342"

	colIMDBID     = 0
	colYourRating = 1
	colDateRated  = 2
	colTitle      = 3
	colURL        = 5
	colYear       = 9
)

type rating struct {
	IMDBID     string
	Title      string
	YourRating int
	Year       int
	URL        string
	DateRated  time.Time
	Poster     image.Image
}

type tmdbPosterResult struct {
	PosterPath string `json:"poster_path"`
}

type tmdbFindResponse struct {
	MovieResults []tmdbPosterResult `json:"movie_results"`
	TVResults    []tmdbPosterResult `json:"tv_results"`
}

func Run(verbose bool) error {
	log := util.Logger("imdb")

	accessToken := os.Getenv(tmdbTokenEnv)
	if accessToken == "" {
		log.Error("environment variable not set", "env", tmdbTokenEnv)
		return fmt.Errorf("%s environment variable not set", tmdbTokenEnv)
	}

	records, err := readCSV("data/imdb.csv")
	if err != nil {
		return err
	}

	log.Info("processing ratings", "count", min(len(records)-1, 5))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	httpClient := util.NewClient()

	var ratings []rating
	for i := 1; i < len(records) && len(ratings) < 5; i++ {
		r, err := parseRating(records[i])
		if err != nil {
			continue
		}

		poster, err := fetchPoster(ctx, httpClient, accessToken, r.IMDBID)
		if err != nil {
			log.Warn("could not fetch poster", "title", r.Title, "err", err)
		} else {
			r.Poster = poster
		}

		ratings = append(ratings, r)
	}

	var posters []image.Image
	for _, r := range ratings {
		if r.Poster != nil {
			posters = append(posters, r.Poster)
		}
	}

	if len(posters) > 0 {
		collagePath, err := util.CreateCollage(posters, util.CollageOptions{
			TileHeight: 300,
			OutputPath: "assets/img/imdb-top-ratings",
		})
		if err != nil {
			return fmt.Errorf("creating collage: %w", err)
		}
		log.Info("collage saved", "path", collagePath)

		state, err := util.LoadState()
		if err != nil {
			return fmt.Errorf("loading state: %w", err)
		}

		state.ImdbImagePath = collagePath

		if err := util.SaveState(state); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}

		if err := util.WriteReadme(state); err != nil {
			return err
		}

		log.Info("readme updated")
	}

	for i, r := range ratings {
		log.Debug("rated", "num", i+1, "title", r.Title, "year", r.Year, "rating", r.YourRating, "date", r.DateRated.Format("2006-01-02"))
	}

	return nil
}

func readCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening CSV file: %w", err)
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading CSV file: %w", err)
	}

	if len(records) <= 1 {
		return nil, fmt.Errorf("no data found in CSV file")
	}

	return records, nil
}

func parseRating(record []string) (rating, error) {
	var r rating

	if v := record[colYourRating]; v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return r, fmt.Errorf("parsing Your Rating: %w", err)
		}
		r.YourRating = n
	}

	if v := record[colDateRated]; v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return r, fmt.Errorf("parsing Date Rated: %w", err)
		}
		r.DateRated = t
	}

	r.IMDBID = record[colIMDBID]
	r.Title = strings.Trim(record[colTitle], "\"")
	r.URL = record[colURL]

	if v := record[colYear]; v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			r.Year = n
		}
	}

	return r, nil
}

func fetchPoster(ctx context.Context, client *http.Client, token, imdbID string) (image.Image, error) {
	posterURL, err := resolvePosterURL(ctx, client, token, imdbID)
	if err != nil {
		return nil, err
	}

	data, err := tmdbGet(ctx, client, token, posterURL)
	if err != nil {
		return nil, fmt.Errorf("downloading poster: %w", err)
	}

	img, err := util.DecodeImage(data)
	if err != nil {
		return nil, fmt.Errorf("decoding poster image: %w", err)
	}

	return img, nil
}

func resolvePosterURL(ctx context.Context, client *http.Client, token, imdbID string) (string, error) {
	findURL := fmt.Sprintf(tmdbFindURL, imdbID)

	data, err := tmdbGet(ctx, client, token, findURL)
	if err != nil {
		return "", fmt.Errorf("searching TMDB: %w", err)
	}

	var result tmdbFindResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if posterPath := firstPosterPath(result); posterPath == "" {
		return "", fmt.Errorf("no poster found for %s", imdbID)
	} else {
		return tmdbImageBase + posterPath, nil
	}
}

func firstPosterPath(r tmdbFindResponse) string {
	for _, m := range r.MovieResults {
		if m.PosterPath != "" {
			return m.PosterPath
		}
	}
	for _, t := range r.TVResults {
		if t.PosterPath != "" {
			return t.PosterPath
		}
	}
	return ""
}

func tmdbGet(ctx context.Context, client *http.Client, token, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
