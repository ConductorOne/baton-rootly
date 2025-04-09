package client

import (
	"context"
	"fmt"
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
	queryParameters map[string]interface{},
	pathParameters ...string,
) *url.URL {
	// query parameters
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
	parsedURL := c.generateURL(path, map[string]interface{}{
		"page[number]": 1,
		"page[size]":   c.resourcesPageSize,
	})
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
// TODO: implement pagination.
func (c *Client) ListScheduleRotations(
	ctx context.Context,
	scheduleID string,
) ([]string, error) {
	logger := ctxzap.Extract(ctx)
	parsedURL := c.generateURL(ListScheduleRotationsAPIEndpoint, map[string]interface{}{
		"page[number]": 1,
		"page[size]":   400, // note: set really high to 400 since we don't have pagination logic
	}, scheduleID)
	logger.Debug("Generated URL", zap.String("parsedURL", parsedURL.String()))

	var resp ScheduleRotationsResponse
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
	if resp.Links.Next != "" {
		logger.Error("Unexpected paginated Next Link", zap.String("resp.Links.Next", resp.Links.Next))
		return nil, fmt.Errorf(
			"schedule rotations exceed maximum supported: %v of %v",
			resp.Meta.TotalCount,
			400,
		)
	}

	var rotationIDs []string
	for _, rotation := range resp.Data {
		if rotation.Type != "schedule_rotations" {
			logger.Debug("Unexpected type in schedule rotation", zap.String("rotation.Type", rotation.Type))
			continue
		}
		rotationIDs = append(rotationIDs, rotation.ID)
	}
	return rotationIDs, nil
}

// ListScheduleRotationUsers returns a list of user IDs for a given schedule rotation ID.
// TODO: implement pagination.
func (c *Client) ListScheduleRotationUsers(
	ctx context.Context,
	rotationID string,
) ([]int, error) {
	logger := ctxzap.Extract(ctx)
	parsedURL := c.generateURL(ListScheduleRotationUsersAPIEndpoint, map[string]interface{}{
		"page[number]": 1,
		"page[size]":   400, // note: set really high to 400 since we don't have pagination logic
	}, rotationID)
	logger.Debug("Generated URL", zap.String("parsedURL", parsedURL.String()))

	var resp ScheduleRotationUsersResponse
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
	if resp.Links.Next != "" {
		logger.Error("Unexpected paginated Next Link", zap.String("resp.Links.Next", resp.Links.Next))
		return nil, fmt.Errorf(
			"schedule rotation users exceed maximum supported: %v of %v",
			resp.Meta.TotalCount,
			400,
		)
	}

	var userIDs []int
	for _, user := range resp.Data {
		userIDs = append(userIDs, user.Attributes.UserID)
	}
	return userIDs, nil
}

// GetScheduleMemberIDs returns a list of member user IDs for a given schedule ID.
func (c *Client) GetScheduleMemberIDs(
	ctx context.Context,
	scheduleID string,
) ([]int, error) {
	logger := ctxzap.Extract(ctx)
	rotationIDs, err := c.ListScheduleRotations(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	if len(rotationIDs) == 0 {
		logger.Debug("No schedule rotations found", zap.String("scheduleID", scheduleID))
		return nil, nil
	}

	var userIDs []int
	for _, rotationID := range rotationIDs {
		rotationUserIDs, err := c.ListScheduleRotationUsers(ctx, rotationID)
		if err != nil {
			return nil, err
		}
		if len(rotationUserIDs) == 0 {
			logger.Debug(
				"No schedule rotation users found",
				zap.String("scheduleID", scheduleID),
				zap.String("rotationID", rotationID),
			)
			continue
		}
		userIDs = append(userIDs, rotationUserIDs...)
	}

	return userIDs, nil
}

// ListOnCallUsers returns a list of on-call user IDs for a given schedule ID.
func (c *Client) ListOnCallUsers(
	ctx context.Context,
	scheduleID string,
) ([]int, error) {
	logger := ctxzap.Extract(ctx)
	now := time.Now()
	parsedURL := c.generateURL(ListScheduleShiftsAPIEndpoint, map[string]interface{}{
		"include":      "user",
		"schedule_ids": fmt.Sprintf(`[%s]`, scheduleID),
		"from":         now.Format(time.RFC3339),
		"to":           now.Add(1 * time.Hour).Format(time.RFC3339),
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
