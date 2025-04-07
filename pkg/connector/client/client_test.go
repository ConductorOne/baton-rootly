package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/stretchr/testify/require"
)

const (
	testAPIKey                    = "test-api-key"
	testBaseURLStr                = "https://api.example.com"
	usersListResultsPage1of2Size1 = `{
    "data": [
        {
            "id": "97487",
            "type": "users",
            "attributes": {
                "name": "Sam Testsalot",
                "email": "sam.testsalot@team1.com",
                "phone": "+12345678910",
                "phone_2": null,
                "full_name": "Sam Testsalot",
                "full_name_with_team": "[team1] Sam Testsalot",
                "slack_id": "@testsalot",
                "time_zone": "America/New_York",
                "updated_at": "2025-04-02T13:38:10.476-07:00",
                "created_at": "2025-03-28T07:05:58.946-07:00"
            }
        }
    ],
    "links": {
        "self": "https://api.rootly.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "first": "https://api.rootly.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "prev": null,
        "next": "https://api.rootly.com/v1/users?page%5Bnumber%5D=2&page%5Bsize%5D=1",
        "last": "https://api.rootly.com/v1/users?page%5Bnumber%5D=2&page%5Bsize%5D=1"
    },
    "meta": {
        "current_page": 1,
        "next_page": 2,
        "prev_page": null,
        "total_count": 2,
        "total_pages": 2
    }
}`
	usersListResultsPage2of2Size1 = `{
    "data": [
		{
            "id": "96913",
            "type": "users",
            "attributes": {
                "name": "Jo Codesalot",
                "email": "jo.codesalot@team1.com",
                "phone": null,
                "phone_2": null,
                "full_name": "Jo Codesalot",
                "full_name_with_team": "[Team1] Jo Codesalot",
                "slack_id": "@codesalot",
                "time_zone": "America/Los_Angeles",
                "updated_at": "2025-04-01T12:10:36.179-07:00",
                "created_at": "2025-04-01T12:09:34.175-07:00"
            }
        }
    ],
    "links": {
        "self": "https://api.rootly.com/v1/users?page%5Bnumber%5D=2&page%5Bsize%5D=1",
        "first": "https://api.rootly.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "prev": "https://api.rootly.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "next": null,
        "last": "https://api.rootly.com/v1/users?page%5Bnumber%5D=2&page%5Bsize%5D=1"
    },
    "meta": {
        "current_page": 2,
        "next_page": null,
        "prev_page": 1,
        "total_count": 2,
        "total_pages": 2
    }
}`
	usersListResultsPage1of1Size2 = `{
    "data": [
        {
            "id": "97487",
            "type": "users",
            "attributes": {
                "name": "Sam Testsalot",
                "email": "sam.testsalot@team1.com",
                "phone": "+12345678910",
                "phone_2": null,
                "full_name": "Sam Testsalot",
                "full_name_with_team": "[Team1] Sam Testsalot",
                "slack_id": "@testsalot",
                "time_zone": "America/New_York",
                "updated_at": "2025-04-02T13:38:10.476-07:00",
                "created_at": "2025-03-28T07:05:58.946-07:00"
            }
        },
		{
            "id": "96913",
            "type": "users",
            "attributes": {
                "name": "Jo Codesalot",
                "email": "jo.codesalot@team1.com",
                "phone": null,
                "phone_2": null,
                "full_name": "Jo Codesalot",
                "full_name_with_team": "[Team1] Jo Codesalot",
                "slack_id": "@codesalot",
                "time_zone": "America/Los_Angeles",
                "updated_at": "2025-04-01T12:10:36.179-07:00",
                "created_at": "2025-04-01T12:09:34.175-07:00"
            }
        }
    ],
    "links": {
        "self": "https://api.rootly.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=2",
        "first": "https://api.rootly.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=2",
        "prev": null,
        "next": null,
        "last": "https://api.rootly.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=2"
    },
    "meta": {
        "current_page": 1,
        "next_page": null,
        "prev_page": null,
        "total_count": 2,
        "total_pages": 1
    }
}`
)

func TestClient_GetUsers(t *testing.T) {
	type fields struct {
		resourcesPageSize int
		responseBody      string
	}
	type args struct {
		// note we're only using the path here for test input purposes
		// since this will get tacked onto the server URL later
		pTokenPath string
	}
	type want struct {
		users       []User
		nextToken   string
		expectError bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "list pagination page 1 of 2, size 1",
			fields: fields{
				resourcesPageSize: 1,
				responseBody:      usersListResultsPage1of2Size1,
			},
			args: args{
				pTokenPath: "",
			},
			want: want{
				users: []User{
					{
						ID:   "97487",
						Type: "users",
						Attributes: UserAttributes{
							Name:      "Sam Testsalot",
							Email:     "sam.testsalot@team1.com",
							Phone:     "+12345678910",
							FullName:  "Sam Testsalot",
							SlackID:   "@testsalot",
							UpdatedAt: "2025-04-02T13:38:10.476-07:00",
							CreatedAt: "2025-03-28T07:05:58.946-07:00",
						},
					},
				},
				nextToken:   "https://api.rootly.com/v1/users?page%5Bnumber%5D=2&page%5Bsize%5D=1",
				expectError: false,
			},
		},
		{
			name: "list pagination page 2 of 2, size 1",
			fields: fields{
				resourcesPageSize: 1,
				responseBody:      usersListResultsPage2of2Size1,
			},
			args: args{
				pTokenPath: "/v1/users?page%5Bnumber%5D=2&page%5Bsize%5D=1",
			},
			want: want{
				users: []User{
					{
						ID:   "96913",
						Type: "users",
						Attributes: UserAttributes{
							Name:      "Jo Codesalot",
							Email:     "jo.codesalot@team1.com",
							Phone:     "",
							FullName:  "Jo Codesalot",
							SlackID:   "@codesalot",
							UpdatedAt: "2025-04-01T12:10:36.179-07:00",
							CreatedAt: "2025-04-01T12:09:34.175-07:00",
						},
					},
				},
				nextToken:   "",
				expectError: false,
			},
		},
		{
			name: "list pagination page 1 of 1, size 2",
			fields: fields{
				resourcesPageSize: 2,
				responseBody:      usersListResultsPage1of1Size2,
			},
			args: args{
				pTokenPath: "",
			},
			want: want{
				users: []User{
					{
						ID:   "97487",
						Type: "users",
						Attributes: UserAttributes{
							Name:      "Sam Testsalot",
							Email:     "sam.testsalot@team1.com",
							Phone:     "+12345678910",
							FullName:  "Sam Testsalot",
							SlackID:   "@testsalot",
							UpdatedAt: "2025-04-02T13:38:10.476-07:00",
							CreatedAt: "2025-03-28T07:05:58.946-07:00",
						},
					},
					{
						ID:   "96913",
						Type: "users",
						Attributes: UserAttributes{
							Name:      "Jo Codesalot",
							Email:     "jo.codesalot@team1.com",
							Phone:     "",
							FullName:  "Jo Codesalot",
							SlackID:   "@codesalot",
							UpdatedAt: "2025-04-01T12:10:36.179-07:00",
							CreatedAt: "2025-04-01T12:09:34.175-07:00",
						},
					},
				},
				nextToken:   "",
				expectError: false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(writer http.ResponseWriter, request *http.Request) {
						writer.Header().Set(uhttp.ContentType, "application/json")
						writer.WriteHeader(http.StatusOK)
						_, err := writer.Write([]byte(tc.fields.responseBody))
						if err != nil {
							return
						}
					},
				),
			)
			defer server.Close()

			ctx := context.Background()
			client, err := NewClient(
				ctx,
				server.URL,
				testAPIKey,
				tc.fields.resourcesPageSize,
			)
			if err != nil {
				t.Fatal(err)
			}

			var pToken string
			if tc.args.pTokenPath != "" {
				// concat the test server URL and pTokenPath
				pToken = server.URL + tc.args.pTokenPath
			}

			users, nextPageToken, err := client.GetUsers(ctx, pToken)
			if tc.want.expectError {
				require.NotNil(t, err, "GetUsers() error = %v, not expecting an error", err)
			} else {
				require.Nil(t, err)
			}

			require.Len(t, users, len(tc.want.users))
			require.ElementsMatch(t, tc.want.users, users)
			require.Equal(t, tc.want.nextToken, nextPageToken)
		})
	}
}

func TestClient_generateURL(t *testing.T) {
	ctx := context.Background()
	client, err := NewClient(ctx, testBaseURLStr, testAPIKey, 2)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		path            string
		queryParameters map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want *url.URL
	}{
		{
			name: "empty path, no query parameters",
			args: args{
				path:            "",
				queryParameters: map[string]interface{}{},
			},
			want: &url.URL{
				Scheme: "https",
				Host:   "api.example.com",
			},
		},
		{
			name: "no query parameters, path starts with backslash",
			args: args{
				path:            "/v1/test",
				queryParameters: map[string]interface{}{},
			},
			want: &url.URL{
				Scheme: "https",
				Host:   "api.example.com",
				Path:   "v1/test",
			},
		},
		{
			name: "no query parameters, path starts without backslash",
			args: args{
				path:            "v1/test",
				queryParameters: map[string]interface{}{},
			},
			want: &url.URL{
				Scheme: "https",
				Host:   "api.example.com",
				Path:   "v1/test",
			},
		},
		{
			name: "single string query parameter",
			args: args{
				path: "/v1/test",
				queryParameters: map[string]interface{}{
					"param1": "value1",
				},
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "v1/test",
				RawQuery: "param1=value1",
			},
		},
		{
			name: "multiple value types as query parameters",
			args: args{
				path: "/v1/test",
				queryParameters: map[string]interface{}{
					"param1": "value1",
					"param2": 123,
					"param3": true,
				},
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "v1/test",
				RawQuery: "param1=value1&param2=123&param3=true",
			},
		},
		{
			name: "skips unsupported value types as query parameter",
			args: args{
				path: "/v1/test",
				queryParameters: map[string]interface{}{
					"param1": "value1",
					"param2": []int{1, 2, 3}, // should skip slice
				},
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "v1/test",
				RawQuery: "param1=value1",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := client.generateURL(tc.args.path, tc.args.queryParameters)
			require.Equal(t, tc.want, got)
		})
	}
}
