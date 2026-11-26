package connector

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/conductorone/baton-rootly/pkg/connector/client"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	sdkResource "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/stretchr/testify/require"
)

func Test_getBestName(t *testing.T) {
	type args struct {
		userAttr client.UserAttributes
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "choose Name when available",
			args: args{
				userAttr: client.UserAttributes{
					Name:     "Sam",
					FullName: "Sam Testalot",
					Email:    "sam.testalot@example.com",
				},
			},
			want: "Sam",
		},
		{
			name: "choose FullName if Name is not available",
			args: args{
				userAttr: client.UserAttributes{
					FullName: "Sam Testalot",
					Email:    "sam.testalot@example.com",
				},
			},
			want: "Sam Testalot",
		},
		{
			name: "choose Email when Name and FullName are not available",
			args: args{
				userAttr: client.UserAttributes{
					Email: "sam.testalot@example.com",
				},
			},
			want: "sam.testalot@example.com",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getBestName(tc.args.userAttr)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_getUserProfile(t *testing.T) {
	type args struct {
		user client.User
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "All fields populated",
			args: args{
				user: client.User{
					ID: "123",
					Attributes: client.UserAttributes{
						CreatedAt: "2023-01-01T00:00:00Z", // not captured in profile
						UpdatedAt: "2023-01-02T00:00:00Z",
						Name:      "Sam",
						FullName:  "Sam Testalot",             // not captured in profile
						Email:     "sam.testalot@example.com", // not captured in profile
						SlackID:   "@testalot",
						Phone:     "123-456-7890",
					},
				},
			},
			want: map[string]interface{}{
				"user_id":    "123",
				"updated_at": "2023-01-02T00:00:00Z",
				"slack_id":   "@testalot",
				"phone":      "123-456-7890",
			},
		},
		{
			name: "Only required fields populated",
			args: args{
				user: client.User{
					ID: "124",
					Attributes: client.UserAttributes{
						CreatedAt: "2023-01-01T00:00:00Z", // not captured in profile
						UpdatedAt: "2023-01-02T00:00:00Z",
						Email:     "sam.testalot@example.com", // not captured in profile
					},
				},
			},
			want: map[string]interface{}{
				"user_id":    "124",
				"updated_at": "2023-01-02T00:00:00Z",
			},
		},
		{
			name: "Optional fields partially populated",
			args: args{
				user: client.User{
					ID: "125",
					Attributes: client.UserAttributes{
						CreatedAt: "2023-01-03T00:00:00Z", // not captured in profile
						UpdatedAt: "2023-01-04T00:00:00Z",
						FullName:  "Sam Testalot",             // not captured in profile
						Email:     "sam.testalot@example.com", // not captured in profile
					},
				},
			},
			want: map[string]interface{}{
				"user_id":    "125",
				"updated_at": "2023-01-04T00:00:00Z",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getUserProfile(tc.args.user)
			require.Equal(t, tc.want, got)
		})
	}
}

const (
	usersWithRolesPage1 = `{
    "data": [
        {
            "id": "97487",
            "type": "users",
            "attributes": {"name": "Sam Testsalot", "email": "sam.testsalot@team1.com"},
            "relationships": {
                "role": {"data": {"id": "role-owner-guid", "type": "roles"}},
                "on_call_role": {"data": {"id": "on-call-role-admin-guid", "type": "on_call_roles"}}
            }
        },
        {
            "id": "96913",
            "type": "users",
            "attributes": {"name": "Jo Codesalot", "email": "jo.codesalot@team1.com"},
            "relationships": {
                "role": {"data": null},
                "on_call_role": {"data": {"id": "on-call-role-observer-guid", "type": "on_call_roles"}}
            }
        }
    ],
    "links": {"self": "", "first": "", "prev": null, "next": "%s/v1/users?page%%5Bnumber%%5D=2&page%%5Bsize%%5D=2", "last": ""},
    "meta": {"current_page": 1, "next_page": 2, "prev_page": null, "total_count": 3, "total_pages": 2}
}`
	usersWithRolesPage2 = `{
    "data": [
        {
            "id": "96914",
            "type": "users",
            "attributes": {"name": "Kim Buildsalot", "email": "kim.buildsalot@team1.com"}
        }
    ],
    "links": {"self": "", "first": "", "prev": "", "next": null, "last": ""},
    "meta": {"current_page": 2, "next_page": null, "prev_page": 1, "total_count": 3, "total_pages": 2}
}`
)

// Test_userBuilder_GrantsForResourceType walks both user pages and checks that every role
// relationship becomes exactly one grant, that a null relationship yields none, and that
// pagination terminates on the last page.
func Test_userBuilder_GrantsForResourceType(t *testing.T) {
	ctx := context.Background()

	var serverURL string
	server := httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				body := fmt.Sprintf(usersWithRolesPage1, serverURL)
				if request.URL.Query().Get("page[number]") == "2" {
					body = usersWithRolesPage2
				}
				writer.Header().Set(uhttp.ContentType, "application/json")
				writer.WriteHeader(http.StatusOK)
				if _, err := writer.Write([]byte(body)); err != nil {
					t.Errorf("writing test response failed: %v", err)
				}
			},
		),
	)
	defer server.Close()
	serverURL = server.URL

	rootlyClient, err := client.NewClient(ctx, server.URL, "test-api-key", 2)
	require.NoError(t, err)
	builder := newUserBuilder(rootlyClient)

	grants, results, err := builder.GrantsForResourceType(ctx, userResourceType.Id, sdkResource.SyncOpAttrs{})
	require.NoError(t, err)
	require.NotEmpty(t, results.NextPageToken, "expected a second page of users")

	gotIDs := make([]string, 0, len(grants))
	for _, g := range grants {
		gotIDs = append(gotIDs, g.GetId())
	}
	require.ElementsMatch(t, []string{
		"role:role-owner-guid:assigned:user:97487",
		"on_call_role:on-call-role-admin-guid:assigned:user:97487",
		"on_call_role:on-call-role-observer-guid:assigned:user:96913",
	}, gotIDs)

	// the grant's entitlement must be the one the role syncer emits for the same role
	require.Equal(t, "role:role-owner-guid:assigned", grants[0].GetEntitlement().GetId())
	require.Equal(t, roleResourceType.Id, grants[0].GetEntitlement().GetResource().GetId().GetResourceType())
	require.Equal(t, userResourceType.Id, grants[0].GetPrincipal().GetId().GetResourceType())

	// second page: no relationships at all, and pagination terminates
	page2Grants, page2Results, err := builder.GrantsForResourceType(
		ctx,
		userResourceType.Id,
		sdkResource.SyncOpAttrs{PageToken: pagination.Token{Token: results.NextPageToken}},
	)
	require.NoError(t, err)
	require.Empty(t, page2Grants)
	require.Equal(t, "", page2Results.NextPageToken)
}

func Test_userBuilder_GrantsForResourceType_wrongResourceType(t *testing.T) {
	_, _, err := newUserBuilder(nil).GrantsForResourceType(context.Background(), "team", sdkResource.SyncOpAttrs{})
	require.Error(t, err)
}
