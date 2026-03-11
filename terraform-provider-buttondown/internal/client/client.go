package client

import (
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.buttondown.com"

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

type Option func(*Client)

func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.BaseURL = url
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		BaseURL: defaultBaseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
