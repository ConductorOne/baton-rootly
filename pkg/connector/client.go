package connector

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type Client struct {
	httpClient *uhttp.BaseHttpClient
	baseURL    string
	apiKey     string
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	httpClient, err := uhttp.NewBaseHttpClientWithContext(ctx, http.DefaultClient)
	if err != nil {
		return nil, err
	}

	// This is preferred over using the regular http.Client directly
	// as it provides automatic rate limiting handling, error wrapping with gRPC status codes,
	// and built-in GET response caching
	return &Client{
		httpClient: httpClient,
		baseURL:    "https://api.rootly.com",
		apiKey:     apiKey,
	}, nil
}

// doRequest
func (c *Client) doRequest(
	ctx context.Context,
	method string,
	path string,
	payload interface{},
	target interface{},
) (*http.Response, error) {
	l := ctxzap.Extract(ctx)

	// create the URL
	parsedURL, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, err
	}

	// create the request
	options := []uhttp.RequestOption{
		uhttp.WithBearerToken(c.apiKey),
	}
	if payload != nil {
		options = append(options, uhttp.WithContentTypeVndHeader())
	}
	req, err := c.httpClient.NewRequest(ctx, method, parsedURL, options...)
	if err != nil {
		return nil, err
	}

	l.Debug("sending request", zap.String("url", parsedURL.String()), zap.String("method", method))

	var doOptions []uhttp.DoOption
	if target != nil {
		doOptions = append(doOptions, uhttp.WithJSONResponse(target))
	}

	// do the request and handle the response
	resp, err := c.httpClient.Do(req, doOptions...)
	if err != nil && resp != nil {
		// There is an error. Log the response.
		defer resp.Body.Close()

		bodyBytes, err0 := io.ReadAll(resp.Body)
		if err0 != nil {
			return nil, err0
		}

		l.Error(
			"doRequest error",
			zap.String("url", req.URL.String()),
			zap.String("method", req.Method),
			zap.String("response.status", resp.Status),
			zap.String("response.body", string(bodyBytes)),
			zap.Error(err),
		)
	}
	return resp, err
}
