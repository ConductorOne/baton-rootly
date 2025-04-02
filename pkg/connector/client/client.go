package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const (
	BaseURLStr       = "https://api.rootly.com"
	UsersAPIEndpoint = "/v1/users"
)

type Client struct {
	httpClient *uhttp.BaseHttpClient
	baseURL    *url.URL
	apiKey     string
}

// NewClient creates a new Rootly client.
func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	httpClient, err := uhttp.NewBaseHttpClientWithContext(ctx, http.DefaultClient)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(BaseURLStr)
	if err != nil {
		return nil, err
	}

	// This is preferred over using the regular http.Client directly
	// as it provides automatic rate limiting handling, error wrapping with gRPC status codes,
	// and built-in GET response caching
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		apiKey:     apiKey,
	}, nil
}

// doRequest is a helper for taking various request inputs, issuing a client request, and handling the response.
// It marshals the response body given a target.
func (c *Client) doRequest(
	ctx context.Context,
	method string,
	url *url.URL,
	requestBody interface{},
	target interface{},
) (string, error) {
	l := ctxzap.Extract(ctx)

	// create the request
	reqOptions := []uhttp.RequestOption{
		uhttp.WithBearerToken(c.apiKey),
		uhttp.WithAcceptVndJSONHeader(),
	}
	if requestBody != nil {
		reqOptions = append(reqOptions, uhttp.WithJSONBody(requestBody))
	}
	req, err := c.httpClient.NewRequest(ctx, method, url, reqOptions...)
	if err != nil {
		return "", err
	}

	// do the request and handle the response
	l.Debug("sending request", zap.String("url", url.String()), zap.String("method", method))
	var respOptions []uhttp.DoOption
	if target != nil {
		respOptions = append(respOptions, uhttp.WithJSONResponse(target))
	}
	resp, err := c.httpClient.Do(req, respOptions...)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}
	return "", nil
}

func (c *Client) generateURL(
	path string,
	queryParameters map[string]interface{},
) *url.URL {
	params := url.Values{}
	for key, valueAny := range queryParameters {
		switch value := valueAny.(type) {
		case string:
			params.Add(key, value)
		case int:
			params.Add(key, strconv.Itoa(value))
		case bool:
			params.Add(key, strconv.FormatBool(value))
		default:
			continue
		}
	}

	output := c.baseURL.JoinPath(path)
	output.RawQuery = params.Encode()
	return output
}

func (c *Client) GetUsers(ctx context.Context, _ *pagination.Token) ([]User, string, error) {
	// TODO(steve) implement with pagination
	var resp UsersResponse
	_, err := c.doRequest(
		ctx,
		http.MethodGet,
		c.generateURL(UsersAPIEndpoint, nil),
		nil,
		&resp,
	)
	if err != nil {
		return nil, "", err
	}
	return resp.Data, "", nil
}
