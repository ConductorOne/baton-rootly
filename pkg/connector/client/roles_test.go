package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/stretchr/testify/require"
)

const (
	rolesListResultsPage1of2Size1 = `{
    "data": [
        {
            "id": "role-owner-guid",
            "type": "roles",
            "attributes": {
                "name": "Owner",
                "slug": "owner",
                "is_deletable": false,
                "is_editable": false,
                "incident_permission_set_id": null,
                "incidents_permissions": ["create", "read", "update", "delete"],
                "created_at": "2025-03-28T07:05:58.946-07:00",
                "updated_at": "2025-04-02T13:38:10.476-07:00"
            }
        }
    ],
    "links": {
        "self": "https://api.example.com/v1/roles?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "first": "https://api.example.com/v1/roles?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "prev": null,
        "next": "https://api.example.com/v1/roles?page%5Bnumber%5D=2&page%5Bsize%5D=1",
        "last": "https://api.example.com/v1/roles?page%5Bnumber%5D=2&page%5Bsize%5D=1"
    },
    "meta": {
        "current_page": 1,
        "next_page": 2,
        "prev_page": null,
        "total_count": 2,
        "total_pages": 2
    }
}`
	rolesListResultsPage2of2Size1 = `{
    "data": [
        {
            "id": "role-observer-guid",
            "type": "roles",
            "attributes": {
                "name": "Observer",
                "slug": "observer",
                "is_deletable": true,
                "is_editable": true,
                "created_at": "2025-03-29T07:05:58.946-07:00",
                "updated_at": "2025-04-03T13:38:10.476-07:00"
            }
        }
    ],
    "links": {
        "self": "https://api.example.com/v1/roles?page%5Bnumber%5D=2&page%5Bsize%5D=1",
        "first": "https://api.example.com/v1/roles?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "prev": "https://api.example.com/v1/roles?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "next": null,
        "last": "https://api.example.com/v1/roles?page%5Bnumber%5D=2&page%5Bsize%5D=1"
    },
    "meta": {
        "current_page": 2,
        "next_page": null,
        "prev_page": 1,
        "total_count": 2,
        "total_pages": 2
    }
}`
	onCallRolesListResultsPage1of1Size2 = `{
    "data": [
        {
            "id": "on-call-role-admin-guid",
            "type": "on_call_roles",
            "attributes": {
                "name": "Admin",
                "slug": "admin",
                "system_role": "admin",
                "schedules_permissions": ["create", "read", "update", "delete"],
                "created_at": "2025-03-28T07:05:58.946-07:00",
                "updated_at": "2025-04-02T13:38:10.476-07:00"
            }
        },
        {
            "id": "on-call-role-admin-2-guid",
            "type": "on_call_roles",
            "attributes": {
                "name": "Admin",
                "slug": "admin-1",
                "system_role": "custom",
                "created_at": "2025-03-29T07:05:58.946-07:00",
                "updated_at": "2025-04-03T13:38:10.476-07:00"
            }
        }
    ],
    "links": {
        "self": "https://api.example.com/v1/on_call_roles?page%5Bnumber%5D=1&page%5Bsize%5D=2",
        "first": "https://api.example.com/v1/on_call_roles?page%5Bnumber%5D=1&page%5Bsize%5D=2",
        "prev": null,
        "next": null,
        "last": "https://api.example.com/v1/on_call_roles?page%5Bnumber%5D=1&page%5Bsize%5D=2"
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

func TestClient_GetRoles(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		pTokenPath    string
		wantRoles     []Role
		wantNextToken string
	}{
		{
			name:         "page 1 of 2 returns the next page link",
			responseBody: rolesListResultsPage1of2Size1,
			pTokenPath:   "",
			wantRoles: []Role{
				{
					ID:   "role-owner-guid",
					Type: "roles",
					Attributes: RoleAttributes{
						Name:        "Owner",
						Slug:        "owner",
						IsDeletable: false,
						IsEditable:  false,
						UpdatedAt:   "2025-04-02T13:38:10.476-07:00",
						CreatedAt:   "2025-03-28T07:05:58.946-07:00",
					},
				},
			},
			wantNextToken: "https://api.example.com/v1/roles?page%5Bnumber%5D=2&page%5Bsize%5D=1",
		},
		{
			name:         "last page terminates pagination with an empty token",
			responseBody: rolesListResultsPage2of2Size1,
			pTokenPath:   "/v1/roles?page%5Bnumber%5D=2&page%5Bsize%5D=1",
			wantRoles: []Role{
				{
					ID:   "role-observer-guid",
					Type: "roles",
					Attributes: RoleAttributes{
						Name:        "Observer",
						Slug:        "observer",
						IsDeletable: true,
						IsEditable:  true,
						UpdatedAt:   "2025-04-03T13:38:10.476-07:00",
						CreatedAt:   "2025-03-29T07:05:58.946-07:00",
					},
				},
			},
			wantNextToken: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := newStaticJSONServer(t, tc.responseBody)
			defer server.Close()

			ctx := context.Background()
			client, err := NewClient(ctx, server.URL, testAPIKey, testPageSize)
			require.NoError(t, err)

			var pToken string
			if tc.pTokenPath != "" {
				pToken = server.URL + tc.pTokenPath
			}

			roles, nextPageToken, err := client.GetRoles(ctx, pToken)
			require.NoError(t, err)
			require.ElementsMatch(t, tc.wantRoles, roles)
			require.Equal(t, tc.wantNextToken, nextPageToken)
		})
	}
}

func TestClient_GetOnCallRoles(t *testing.T) {
	expectedRoles := []OnCallRole{
		{
			ID:   "on-call-role-admin-guid",
			Type: "on_call_roles",
			Attributes: OnCallRoleAttributes{
				Name:       "Admin",
				Slug:       "admin",
				SystemRole: "admin",
				UpdatedAt:  "2025-04-02T13:38:10.476-07:00",
				CreatedAt:  "2025-03-28T07:05:58.946-07:00",
			},
		},
		{
			ID:   "on-call-role-admin-2-guid",
			Type: "on_call_roles",
			Attributes: OnCallRoleAttributes{
				Name:       "Admin",
				Slug:       "admin-1",
				SystemRole: "custom",
				UpdatedAt:  "2025-04-03T13:38:10.476-07:00",
				CreatedAt:  "2025-03-29T07:05:58.946-07:00",
			},
		},
	}

	server := newStaticJSONServer(t, onCallRolesListResultsPage1of1Size2)
	defer server.Close()

	ctx := context.Background()
	client, err := NewClient(ctx, server.URL, testAPIKey, 2)
	require.NoError(t, err)

	roles, nextPageToken, err := client.GetOnCallRoles(ctx, "")
	require.NoError(t, err)
	require.ElementsMatch(t, expectedRoles, roles)
	require.Equal(t, "", nextPageToken)
}

// newStaticJSONServer returns a test server that answers every request with the same JSON body.
func newStaticJSONServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set(uhttp.ContentType, "application/json")
				writer.WriteHeader(http.StatusOK)
				if _, err := writer.Write([]byte(body)); err != nil {
					t.Errorf("writing test response failed: %v", err)
				}
			},
		),
	)
}
