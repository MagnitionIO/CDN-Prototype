package origin

import (
	"context"
	"fmt"
	"net/http"
)

type Client struct {
	// Endpoint string // example: https://:8080
	Client *http.Client
}

func (c *Client) SayHello(ctx context.Context, endpoint string) (*StringResponse, error) {
	if c.Client == nil {
		return nil, ErrNilHttpClient
	}

	resp, err := c.Client.Get(endpoint + "/origin")
	if err != nil {
		return nil, err
	}

	return &StringResponse{Response: resp}, nil
}

func (c *Client) GetObject(ctx context.Context, id int, size int, endpoint string, headers map[string]string) (*ObjectResponse, error) {
	if c.Client == nil {
		return nil, ErrNilHttpClient
	}

	url := fmt.Sprintf("%s/origin/objects/%d?size=%d", endpoint, id, size)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return &ObjectResponse{
		StringResponse: StringResponse{
			Response: resp,
		},
		ID:   id,
		Size: size,
	}, nil
}
