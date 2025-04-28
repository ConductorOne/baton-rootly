package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const (
	BaseURLStr                           = "https://api.rootly.com"
	ListUsersAPIEndpoint                 = "/v1/users"
	ListTeamsAPIEndpoint                 = "/v1/teams"
	GetTeamAPIEndpoint                   = "/v1/teams/%s"
	ListSecretsAPIEndpoint               = "/v1/secrets"
	ListSchedulesAPIEndpoint             = "/v1/schedules"
	GetScheduleAPIEndpoint               = "/v1/schedules/%s"
	ListScheduleRotationsAPIEndpoint     = "/v1/schedules/%s/schedule_rotations"
	ListScheduleRotationUsersAPIEndpoint = "/v1/schedule_rotations/%s/schedule_rotation_users"
	ListScheduleShiftsAPIEndpoint        = "/v1/shifts"
	ResourcesPageSize                    = 200
)

type Client struct {
	httpClient        *uhttp.BaseHttpClient
	baseURL           *url.URL
	apiKey            string
	resourcesPageSize int
}

// NewClient creates a new Rootly client. Allows for a configurable base URL, API key, and resources page size.
func NewClient(ctx context.Context, baseURL string, apiKey string, resourcesPageSize int) (*Client, error) {
	httpClient, err := uhttp.NewBaseHttpClientWithContext(ctx, http.DefaultClient)
	if err != nil {
		return nil, err
	}

	// set a default base URL if none is provided
	if baseURL == "" {
		baseURL = BaseURLStr
	}
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	// set a default resources page size if none is provided
	if resourcesPageSize <= 0 {
		resourcesPageSize = ResourcesPageSize
	}

	// This is preferred over using the regular http.Client directly
	// as it provides automatic rate limiting handling, error wrapping with gRPC status codes,
	// and built-in GET response caching
	return &Client{
		httpClient:        httpClient,
		baseURL:           parsedURL,
		apiKey:            apiKey,
		resourcesPageSize: resourcesPageSize,
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
) error {
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
		return err
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
		return err
	}
	return nil
}

func (c *Client) generateURL(
	path string,
	queryParameters map[string]string,
	pathParameters ...string,
) *url.URL {
	// query parameters
	params := url.Values{}
	for key, value := range queryParameters {
		params.Add(key, value)
	}

	// path parameters
	for _, param := range pathParameters {
		// not perfect, but less error-prone than using fmt.Sprintf which takes any value type and also fails silently
		path = strings.Replace(path, "%s", param, 1)
	}
	output := c.baseURL.JoinPath(path)
	output.RawQuery = params.Encode()
	return output
}

// generateCurrentPaginatedURL either parses the URL from the page token, or generates a new URL
// with initial pagination if there's no token.
func (c *Client) generateCurrentPaginatedURL(
	ctx context.Context,
	pToken string,
	path string,
	pathParameters ...string,
) (*url.URL, error) {
	logger := ctxzap.Extract(ctx)
	if pToken != "" {
		// this is not the first request to this endpoint
		// use the token string, ie a full URL with params already populated based on the prior request
		parsedURL, err := url.Parse(pToken)
		if err != nil {
			return nil, err
		}
		logger.Debug("Parsed token for paginated URL", zap.String("parsedURL", parsedURL.String()))
		return parsedURL, nil
	}

	// otherwise this is the first paginated request to this endpoint
	parsedURL := c.generateURL(
		path,
		map[string]string{
			"page[number]": "1",
			"page[size]":   strconv.Itoa(c.resourcesPageSize),
		},
		pathParameters...,
	)
	logger.Debug("Generated first paginated URL", zap.String("parsedURL", parsedURL.String()))
	return parsedURL, nil
}

// GetUsers fetches users from the Rootly API. It supports pagination using a page token.
func (c *Client) GetUsers(ctx context.Context, pToken string) ([]User, string, error) {
	logger := ctxzap.Extract(ctx)
	parsedURL, err := c.generateCurrentPaginatedURL(ctx, pToken, ListUsersAPIEndpoint)
	if err != nil {
		return nil, "", err
	}

	var resp UsersResponse
	err = c.doRequest(
		ctx,
		http.MethodGet,
		parsedURL,
		nil,
		&resp,
	)
	if err != nil {
		return nil, "", err
	}
	logger.Debug("Paginated URL for the next request", zap.String("resp.Links.Next", resp.Links.Next))
	return resp.Data, resp.Links.Next, nil
}

// GetTeams fetches the teams from the Rootly API. It supports pagination using a page token.
func (c *Client) GetTeams(ctx context.Context, pToken string) ([]Team, string, error) {
	logger := ctxzap.Extract(ctx)
	parsedURL, err := c.generateCurrentPaginatedURL(ctx, pToken, ListTeamsAPIEndpoint)
	if err != nil {
		return nil, "", err
	}

	var resp TeamsResponse
	err = c.doRequest(
		ctx,
		http.MethodGet,
		parsedURL,
		nil,
		&resp,
	)
	if err != nil {
		return nil, "", err
	}
	logger.Debug("Paginated URL for the next request", zap.String("resp.Links.Next", resp.Links.Next))
	return resp.Data, resp.Links.Next, nil
}

// GetTeamMemberAndAdminIDs returns a list of member user IDs and admin user IDs for a given team ID.
func (c *Client) GetTeamMemberAndAdminIDs(
	ctx context.Context,
	teamID string,
) ([]int, []int, error) {
	logger := ctxzap.Extract(ctx)
	parsedURL := c.generateURL(GetTeamAPIEndpoint, nil, teamID)
	logger.Debug("Generated URL", zap.String("parsedURL", parsedURL.String()))

	var resp TeamResponse
	err := c.doRequest(
		ctx,
		http.MethodGet,
		parsedURL,
		nil,
		&resp,
	)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data.Attributes.UserIDs, resp.Data.Attributes.AdminIDs, nil
}

// GetSecrets fetches the secrets from the Rootly API. It supports pagination using a page token.
func (c *Client) GetSecrets(ctx context.Context, pToken string) ([]Secret, string, error) {
	logger := ctxzap.Extract(ctx)
	parsedURL, err := c.generateCurrentPaginatedURL(ctx, pToken, ListSecretsAPIEndpoint)
	if err != nil {
		return nil, "", err
	}

	var resp SecretsResponse
	err = c.doRequest(
		ctx,
		http.MethodGet,
		parsedURL,
		nil,
		&resp,
	)
	if err != nil {
		return nil, "", err
	}
	logger.Debug("Paginated URL for the next request", zap.String("resp.Links.Next", resp.Links.Next))
	return resp.Data, resp.Links.Next, nil
}

// GetSchedules fetches the schedules from the Rootly API. It supports pagination using a page token.
func (c *Client) GetSchedules(ctx context.Context, pToken string) ([]Schedule, string, error) {
	logger := ctxzap.Extract(ctx)
	parsedURL, err := c.generateCurrentPaginatedURL(ctx, pToken, ListSchedulesAPIEndpoint)
	if err != nil {
		return nil, "", err
	}

	var resp SchedulesResponse
	err = c.doRequest(
		ctx,
		http.MethodGet,
		parsedURL,
		nil,
		&resp,
	)
	if err != nil {
		return nil, "", err
	}
	logger.Debug("Paginated URL for the next request", zap.String("resp.Links.Next", resp.Links.Next))
	return resp.Data, resp.Links.Next, nil
}

// GetScheduleOwnerIDs returns an owner user ID and a list of owner team IDs for a given schedule ID.
func (c *Client) GetScheduleOwnerIDs(
	ctx context.Context,
	scheduleID string,
) (*int, []string, error) {
	logger := ctxzap.Extract(ctx)
	parsedURL := c.generateURL(GetScheduleAPIEndpoint, nil, scheduleID)
	logger.Debug("Generated URL", zap.String("parsedURL", parsedURL.String()))

	var resp ScheduleResponse
	err := c.doRequest(
		ctx,
		http.MethodGet,
		parsedURL,
		nil,
		&resp,
	)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data.Attributes.OwnerUserID, resp.Data.Attributes.OwnerGroupIDs, nil
}

// ListScheduleRotations returns a list of schedule rotation IDs for a given schedule ID.
// It supports pagination using a page token.
func (c *Client) ListScheduleRotations(
	ctx context.Context,
	scheduleID string,
	pToken string,
) ([]string, string, error) {
	logger := ctxzap.Extract(ctx)
	parsedURL, err := c.generateCurrentPaginatedURL(ctx, pToken, ListScheduleRotationsAPIEndpoint, scheduleID)
	if err != nil {
		return nil, "", err
	}
	logger.Debug("Generated URL", zap.String("parsedURL", parsedURL.String()))

	var resp ScheduleRotationsResponse
	err = c.doRequest(
		ctx,
		http.MethodGet,
		parsedURL,
		nil,
		&resp,
	)
	if err != nil {
		return nil, "", err
	}

	var rotationIDs []string
	for _, rotation := range resp.Data {
		if rotation.Type != "schedule_rotations" {
			logger.Debug("Unexpected type in schedule rotation", zap.String("rotation.Type", rotation.Type))
			continue
		}
		rotationIDs = append(rotationIDs, rotation.ID)
	}
	logger.Debug("Paginated URL for the next request", zap.String("resp.Links.Next", resp.Links.Next))
	return rotationIDs, resp.Links.Next, nil
}

// ListScheduleRotationUsers returns a list of user IDs for a given schedule rotation ID.
// It supports pagination using a page token.
func (c *Client) ListScheduleRotationUsers(
	ctx context.Context,
	rotationID string,
	pToken string,
) ([]int, string, error) {
	logger := ctxzap.Extract(ctx)
	parsedURL, err := c.generateCurrentPaginatedURL(ctx, pToken, ListScheduleRotationUsersAPIEndpoint, rotationID)
	if err != nil {
		return nil, "", err
	}
	logger.Debug("Generated URL", zap.String("parsedURL", parsedURL.String()))

	var resp ScheduleRotationUsersResponse
	err = c.doRequest(
		ctx,
		http.MethodGet,
		parsedURL,
		nil,
		&resp,
	)
	if err != nil {
		return nil, "", err
	}

	var userIDs []int
	for _, user := range resp.Data {
		userIDs = append(userIDs, user.Attributes.UserID)
	}
	logger.Debug("Paginated URL for the next request", zap.String("resp.Links.Next", resp.Links.Next))
	return userIDs, resp.Links.Next, nil
}

// ListAllScheduleRotationUsers returns a list of all the member user IDs for a given schedule rotation ID.
// It uses pagination under the hood to make one or more requests to build the full list.
func (c *Client) ListAllScheduleRotationUsers(
	ctx context.Context,
	rotationID string,
) ([]int, error) {
	logger := ctxzap.Extract(ctx)
	var userIDs []int
	var currentPage string
	for {
		memberIDs, nextPage, err := c.ListScheduleRotationUsers(ctx, rotationID, currentPage)
		if err != nil {
			return nil, err
		}

		logger.Debug(
			"Schedule rotations users",
			zap.Int("number of memberIDs", len(memberIDs)),
			zap.String("nextPage", nextPage),
		)

		currentPage = nextPage
		userIDs = append(userIDs, memberIDs...)

		if currentPage == "" {
			break
		}
	}

	return userIDs, nil
}

// ListOnCallUsers returns a list of on-call user IDs for a given schedule ID.
func (c *Client) ListOnCallUsers(
	ctx context.Context,
	scheduleID string,
) ([]int, error) {
	logger := ctxzap.Extract(ctx)
	now := time.Now().UTC()
	parsedURL := c.generateURL(ListScheduleShiftsAPIEndpoint, map[string]string{
		"include":        "user",
		"schedule_ids[]": scheduleID,
		"from":           now.Format(time.RFC3339),
		"to":             now.Add(1 * time.Hour).Format(time.RFC3339),
	}, scheduleID)
	logger.Debug("Generated URL", zap.String("parsedURL", parsedURL.String()))

	var resp ScheduleShiftsResponse
	err := c.doRequest(
		ctx,
		http.MethodGet,
		parsedURL,
		nil,
		&resp,
	)
	if err != nil {
		return nil, err
	}

	var userIDs []int
	for _, user := range resp.Included {
		if user.Type != "users" {
			logger.Debug("Unexpected type in on-call included users", zap.String("user.Type", user.Type))
			continue
		}
		userID, err := strconv.Atoi(user.ID)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}
