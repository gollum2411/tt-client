package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"sync"
)

type Client struct {
	baseURL string
	mtx     sync.Mutex
	c       http.Client

	defaultHeaders map[string][]string
}

func NewClient(opts ...ClientOpt) *Client {
	c := Client{}
	for _, opt := range opts {
		opt(&c)
	}

	c.c = http.Client{
		// TODO: timeout and maybe other opts?
	}
	return &c
}

type ClientError struct {
	StatusCode int
	Status     string
}

func (e ClientError) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Status)
}

type client struct {
}

type RequestDetails struct {
	Method       string
	URI          string
	Qps          url.Values
	ExtraHeaders http.Header
	Body         io.ReadCloser

	ResponseOut any
}

func (c *Client) DoRequestAndParse(ctx context.Context, details *RequestDetails) (err error) {
	if c.mtx.TryLock() {
		panic("dev error: forgot to lock client mutex")
	}

	path, err := url.JoinPath(c.baseURL, details.URI)
	if err != nil {
		return fmt.Errorf("failed to build url: %w", err)
	}
	u, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	if details.Qps != nil {
		u.RawQuery = details.Qps.Encode()
	}

	headers := c.makeDefaultHeaders()

	maps.Copy(headers, details.ExtraHeaders)

	req, err := http.NewRequestWithContext(
		ctx,
		details.Method,
		u.String(),
		details.Body,
	)
	req.Header = headers
	req.Body = details.Body

	req.Header = headers
	if err != nil {
		return fmt.Errorf("failed to create http request")
	}

	resp, err := c.c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode < 200 && resp.StatusCode > 299 {
		return ClientError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			// TODO: attempt to parse response body
		}
	}

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(details.ResponseOut)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (c *Client) makeDefaultHeaders() http.Header {
	return c.defaultHeaders
}
