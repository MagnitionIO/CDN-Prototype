package origin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
)

type Client struct {
	// Endpoint string // example: https://:8080
	Client *http.Client
	Logger *zerolog.Logger
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

func (c *Client) GetObject(ctx context.Context, id string, size int, endpoint string, headers map[string]string) (*ObjectResponse, error) {
	if c.Client == nil {
		return nil, ErrNilHttpClient
	}

	url := fmt.Sprintf("%s/origin/objects/%s?size=%d", endpoint, id, size)
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

func (c *Client) DeleteObject(ctx context.Context, id string, size int, endpoint string, headers map[string]string) (*ObjectResponse, error) {
	if c.Client == nil {
		return nil, ErrNilHttpClient
	}

	url := fmt.Sprintf("%s/origin/objects/%s?size=%d", endpoint, id, size)
	req, err := http.NewRequest("PURGE", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// requestDump, err_dump := httputil.DumpRequestOut(req, true)
	// if err_dump != nil {
	// 	c.Logger.Err(err_dump).Msg("Error dumping request")
	// }

	// c.Logger.Debug().
	// 	Str("requestDump: ", string(requestDump)).
	// 	Msg("Delete Request")

	// c.Logger.Debug().
	// 	Str("id", id).
	// 	Int("size", size).
	// 	Str("endpoint", endpoint).
	// 	Msg("Deleting Request")

	resp, err := c.Client.Do(req)

	// responseDump, _ := httputil.DumpResponse(resp, true)
	// c.Logger.Debug().
	// 	Str("responseDump: ", string(responseDump)).
	// 	Msg("Delete Response")

	if err != nil {
		return nil, err
	}

	// c.Logger.Debug().
	// 	Str("id", id).
	// 	Int("size", size).
	// 	Str("endpoint", endpoint).
	// 	Msg("Finished Deleting Request")

	return &ObjectResponse{
		StringResponse: StringResponse{
			Response: resp,
		},
		ID:   id,
		Size: size,
	}, nil
}

func (c *Client) EnforceL2Object(ctx context.Context, id string, size int, endpoint string, headers map[string]string) (*ObjectResponse, error) {
	if c.Client == nil {
		return nil, ErrNilHttpClient
	}

	url := fmt.Sprintf("%s/origin/objects/enforce-l2/%s?size=%d", endpoint, id, size)
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
