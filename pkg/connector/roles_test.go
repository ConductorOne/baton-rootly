package connector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/conductorone/baton-rootly/pkg/connector/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	sdkResource "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/stretchr/testify/require"
)

const rolesListResponse = `{
    "data": [
        {
            "id": "role-owner-guid",
            "type": "roles",
            "attributes": {
                "name": "Owner",
                "slug": "owner",
                "is_deletable": false,
                "is_editable": false,
                "created_at": "2025-03-28T07:05:58.946-07:00",
                "updated_at": "2025-04-02T13:38:10.476-07:00"
            }
        }
    ],
    "links": {"self": "", "first": "", "prev": null, "next": null, "last": ""},
    "meta": {"current_page": 1, "next_page": null, "prev_page": null, "total_count": 1, "total_pages": 1}
}`

const onCallRolesListResponse = `{
    "data": [
        {
            "id": "on-call-role-admin-guid",
            "type": "on_call_roles",
            "attributes": {
                "name": "Admin",
                "slug": "admin-1",
                "system_role": "admin",
                "created_at": "2025-03-28T07:05:58.946-07:00",
                "updated_at": "2025-04-02T13:38:10.476-07:00"
            }
        }
    ],
    "links": {"self": "", "first": "", "prev": null, "next": null, "last": ""},
    "meta": {"current_page": 1, "next_page": null, "prev_page": null, "total_count": 1, "total_pages": 1}
}`

func Test_roleDisplayName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		slug string
		want string
	}{
		{
			name: "slug that only restates the name is left off",
			in:   "Admin",
			slug: "admin",
			want: "Admin",
		},
		{
			name: "multi-word name matching its kebab-case slug is left alone",
			in:   "Incident Commander",
			slug: "incident-commander",
			want: "Incident Commander",
		},
		{
			name: "disambiguating slug is appended",
			in:   "Admin",
			slug: "admin-1",
			want: "Admin (admin-1)",
		},
		{
			name: "missing slug falls back to the name",
			in:   "Admin",
			slug: "",
			want: "Admin",
		},
		{
			name: "missing name falls back to the slug",
			in:   "",
			slug: "admin-1",
			want: "admin-1",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, roleDisplayName(tc.in, tc.slug))
		})
	}
}

func Test_roleBuilder_List(t *testing.T) {
	server := newConnectorTestServer(t, map[string]string{"/v1/roles": rolesListResponse})
	defer server.Close()

	ctx := context.Background()
	rootlyClient, err := client.NewClient(ctx, server.URL, "test-api-key", 200)
	require.NoError(t, err)

	resources, nextPage, _, err := newRoleBuilder(rootlyClient).List(ctx, nil, &pagination.Token{})
	require.NoError(t, err)
	require.Equal(t, "", nextPage)
	require.Len(t, resources, 1)

	role := resources[0]
	require.Equal(t, "Owner", role.GetDisplayName())
	require.Equal(t, roleResourceType.Id, role.GetId().GetResourceType())
	require.Equal(t, "role-owner-guid", role.GetId().GetResource())

	profile := sdkResource.GetProfile(role).AsMap()
	require.Equal(t, "role-owner-guid", profile["role_id"])
	require.Equal(t, "Owner", profile["name"])
	require.Equal(t, "owner", profile["slug"])
}

func Test_roleBuilder_Entitlements(t *testing.T) {
	ctx := context.Background()
	resource := &v2.Resource{
		Id:          &v2.ResourceId{ResourceType: roleResourceType.Id, Resource: "role-owner-guid"},
		DisplayName: "Owner",
	}

	entitlements, nextPage, _, err := newRoleBuilder(nil).Entitlements(ctx, resource, &pagination.Token{})
	require.NoError(t, err)
	require.Equal(t, "", nextPage)
	require.Len(t, entitlements, 1)
	require.Equal(t, "role:role-owner-guid:assigned", entitlements[0].GetId())
	require.Equal(t, roleAssignedEntitlement, entitlements[0].GetSlug())
	require.Equal(t, "Owner Assigned", entitlements[0].GetDisplayName())
}

// Test_roleBuilder_Grants pins the split: role grants come from the user syncer, not from here.
func Test_roleBuilder_Grants(t *testing.T) {
	ctx := context.Background()
	resource := &v2.Resource{
		Id: &v2.ResourceId{ResourceType: roleResourceType.Id, Resource: "role-owner-guid"},
	}

	grants, nextPage, _, err := newRoleBuilder(nil).Grants(ctx, resource, &pagination.Token{})
	require.NoError(t, err)
	require.Equal(t, "", nextPage)
	require.Empty(t, grants)
}

func Test_onCallRoleBuilder_List(t *testing.T) {
	server := newConnectorTestServer(t, map[string]string{"/v1/on_call_roles": onCallRolesListResponse})
	defer server.Close()

	ctx := context.Background()
	rootlyClient, err := client.NewClient(ctx, server.URL, "test-api-key", 200)
	require.NoError(t, err)

	resources, nextPage, _, err := newOnCallRoleBuilder(rootlyClient).List(ctx, nil, &pagination.Token{})
	require.NoError(t, err)
	require.Equal(t, "", nextPage)
	require.Len(t, resources, 1)

	role := resources[0]
	// the slug disambiguates this Admin from the other per-team Admin roles Rootly creates
	require.Equal(t, "Admin (admin-1)", role.GetDisplayName())
	require.Equal(t, onCallRoleResourceType.Id, role.GetId().GetResourceType())
	require.Equal(t, "on-call-role-admin-guid", role.GetId().GetResource())

	profile := sdkResource.GetProfile(role).AsMap()
	require.Equal(t, "on-call-role-admin-guid", profile["on_call_role_id"])
	require.Equal(t, "admin", profile["system_role"])
}

func Test_onCallRoleBuilder_Entitlements(t *testing.T) {
	ctx := context.Background()
	resource := &v2.Resource{
		Id:          &v2.ResourceId{ResourceType: onCallRoleResourceType.Id, Resource: "on-call-role-admin-guid"},
		DisplayName: "Admin (admin-1)",
	}

	entitlements, nextPage, _, err := newOnCallRoleBuilder(nil).Entitlements(ctx, resource, &pagination.Token{})
	require.NoError(t, err)
	require.Equal(t, "", nextPage)
	require.Len(t, entitlements, 1)
	require.Equal(t, "on_call_role:on-call-role-admin-guid:assigned", entitlements[0].GetId())
	require.Equal(t, roleAssignedEntitlement, entitlements[0].GetSlug())
}

// newConnectorTestServer serves a canned JSON body per request path.
func newConnectorTestServer(t *testing.T, bodiesByPath map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				body, ok := bodiesByPath[request.URL.Path]
				if !ok {
					t.Errorf("unexpected request path %q", request.URL.Path)
					writer.WriteHeader(http.StatusNotFound)
					return
				}
				writer.Header().Set(uhttp.ContentType, "application/json")
				writer.WriteHeader(http.StatusOK)
				if _, err := writer.Write([]byte(body)); err != nil {
					t.Errorf("writing test response failed: %v", err)
				}
			},
		),
	)
}
