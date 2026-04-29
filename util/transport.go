package util

import (
	"log/slog"
	"net/http"
	"time"
)

type loggingTransport struct {
	inner http.RoundTripper
	log   *slog.Logger
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.log.Debug("http request",
		"method", req.Method,
		"url", req.URL.String(),
	)

	start := time.Now()
	resp, err := t.inner.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		t.log.Error("http error",
			"method", req.Method,
			"url", req.URL.String(),
			"duration", duration,
			"err", err,
		)
		return nil, err
	}

	t.log.Debug("http response",
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.StatusCode,
		"duration", duration,
	)

	return resp, nil
}
