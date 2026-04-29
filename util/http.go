package util

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

const (
	defaultTimeout             = 30 * time.Second
	defaultDialTimeout         = 5 * time.Second
	defaultTLSHandshakeTimeout = 5 * time.Second
	defaultIdleConnTimeout     = 30 * time.Second
	defaultMaxIdleConns        = 10
	defaultMaxIdleConnsPerHost = 5
	userAgent                  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"
)

func NewClient() *http.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   defaultDialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
		TLSHandshakeTimeout:   defaultTLSHandshakeTimeout,
		IdleConnTimeout:       defaultIdleConnTimeout,
		MaxIdleConns:          defaultMaxIdleConns,
		MaxIdleConnsPerHost:   defaultMaxIdleConnsPerHost,
		ResponseHeaderTimeout: defaultTimeout,
	}

	return &http.Client{
		Timeout: defaultTimeout,
		Transport: &loggingTransport{
			inner: transport,
			log:   Logger("http"),
		},
	}
}

func setUserAgent(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
}
