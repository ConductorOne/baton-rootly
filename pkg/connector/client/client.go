package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/google/jsonapi"
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

// TODO: move into baton-sdk.
func withVndApiJSONBody(body interface{}) uhttp.RequestOption {
	return func() (io.ReadWriter, map[string]string, error) {
		buffer := new(bytes.Buffer)
		err := jsonapi.MarshalPayload(buffer, body)
		if err != nil {
			return nil, nil, err
		}

		_, headers, err := uhttp.WithContentTypeVndHeader()()
		if err != nil {
			return nil, nil, err
		}

		return buffer, headers, nil
	}
}

// TODO: move into baton-sdk.
func isVndApiJSONContentType(contentType string) bool {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	return strings.HasPrefix(contentType, "application") &&
		strings.Contains(contentType, "vnd.api+json")
}

// TODO: move into baton-sdk.
func withVndApiJSONResponse(response interface{}) uhttp.DoOption {
	return func(resp *uhttp.WrapperResponse) error {
		contentHeader := resp.Header.Get("Content-Type")

		if !isVndApiJSONContentType(contentHeader) {
			if len(resp.Body) != 0 {
				// to print the response, set the envvar BATON_DEBUG_PRINT_RESPONSE_BODY as non-empty, instead
				return fmt.Errorf("unexpected content type for JSON response: %s. status code: %d", contentHeader, resp.StatusCode)
			}
			return fmt.Errorf("unexpected content type for JSON response: %s. status code: %d", contentHeader, resp.StatusCode)
		}
		if response == nil && len(resp.Body) == 0 {
			return nil
		}

		// the sdk wraps resp.Body in a no-op closer before passing into the options, so create a new buffer
		err := jsonapi.UnmarshalPayload(bytes.NewBuffer(resp.Body), response)
		if err != nil {
			// to print the response, set the envvar BATON_DEBUG_PRINT_RESPONSE_BODY as non-empty, instead
			return fmt.Errorf("failed to unmarshal json response: %w. status code: %d", err, resp.StatusCode)
		}
		return nil
	}
}

// doRequest is a helper for taking various request inputs, issuing a JSON:API client request, and handling
// the response. It does not handle any response body unmarshalling, instead leaving that to the caller,
// given the signature differences between jsonapi.UnmarshalPayload and jsonapi.UnmarshalManyPayload.
func (c *Client) doRequest(
	ctx context.Context,
	method string,
	url *url.URL,
	payload interface{},
) (*http.Response, string, error) {
	l := ctxzap.Extract(ctx)

	// create the request
	reqOptions := []uhttp.RequestOption{
		uhttp.WithBearerToken(c.apiKey),
		uhttp.WithAcceptVndJSONHeader(),
	}
	if payload != nil {
		reqOptions = append(reqOptions, withVndApiJSONBody(payload))
	}
	req, err := c.httpClient.NewRequest(ctx, method, url, reqOptions...)
	if err != nil {
		return nil, "", err
	}

	// do the request and handle the response
	l.Debug("sending request", zap.String("url", url.String()), zap.String("method", method))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	return resp, "", nil
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
	resp, _, err := c.doRequest(
		ctx,
		http.MethodGet,
		c.generateURL(UsersAPIEndpoint, nil),
		nil,
	)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	userType := reflect.TypeOf((*User)(nil)).Elem()
	l := ctxzap.Extract(ctx)
	l.Debug(userType.String())
	users, err := jsonapi.UnmarshalManyPayload(resp.Body, userType)
	if err != nil {
		// to print the response, set the envvar BATON_DEBUG_PRINT_RESPONSE_BODY as non-empty, instead
		return nil, "", fmt.Errorf("failed to unmarshal json response: %w. status code: %d", err, resp.StatusCode)
	}
	var typedUsers []User
	for _, user := range users {
		l.Debug("user: ", zap.String("user", fmt.Sprintf("%v", user)))
		l.Debug("user: ", zap.String("user", fmt.Sprintf("%v", user.(User))))
		typedUsers = append(typedUsers, user.(User))
	}
	return typedUsers, "", nil
}
