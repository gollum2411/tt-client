package client

type ClientOpt func(c *Client)

func WithBaseURL(u string) ClientOpt {
	return func(c *Client) {
		c.baseURL = u
	}
}

func WithDefaultHeaders(h map[string][]string) ClientOpt {
	return func(c *Client) {
		c.defaultHeaders = h
	}
}
