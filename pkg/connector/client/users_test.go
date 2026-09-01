package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

const usersListWithRelationships = `{
    "data": [
        {
            "id": "97487",
            "type": "users",
            "attributes": {
                "name": "Sam Testsalot",
                "email": "sam.testsalot@team1.com",
                "full_name": "Sam Testsalot",
                "updated_at": "2025-04-02T13:38:10.476-07:00",
                "created_at": "2025-03-28T07:05:58.946-07:00"
            },
            "relationships": {
                "role": {
                    "data": {
                        "id": "role-owner-guid",
                        "type": "roles"
                    }
                },
                "on_call_role": {
                    "data": {
                        "id": "on-call-role-admin-guid",
                        "type": "on_call_roles"
                    }
                }
            }
        },
        {
            "id": "96913",
            "type": "users",
            "attributes": {
                "name": "Jo Codesalot",
                "email": "jo.codesalot@team1.com",
                "full_name": "Jo Codesalot",
                "updated_at": "2025-04-01T12:10:36.179-07:00",
                "created_at": "2025-04-01T12:09:34.175-07:00"
            },
            "relationships": {
                "role": {
                    "data": null
                },
                "on_call_role": {
                    "data": {
                        "id": "on-call-role-observer-guid",
                        "type": "on_call_roles"
                    }
                }
            }
        },
        {
            "id": "96914",
            "type": "users",
            "attributes": {
                "name": "Kim Buildsalot",
                "email": "kim.buildsalot@team1.com",
                "updated_at": "2025-04-01T12:10:36.179-07:00",
                "created_at": "2025-04-01T12:09:34.175-07:00"
            }
        }
    ],
    "links": {
        "self": "https://api.example.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=3",
        "first": "https://api.example.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=3",
        "prev": null,
        "next": null,
        "last": "https://api.example.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=3"
    },
    "meta": {
        "current_page": 1,
        "next_page": null,
        "prev_page": null,
        "total_count": 3,
        "total_pages": 1
    }
}`

// TestClient_GetUsers_Relationships covers the role relationships Rootly returns inline on the
// user list endpoint: both set, one null, and the relationships object missing entirely.
func TestClient_GetUsers_Relationships(t *testing.T) {
	server := newStaticJSONServer(t, usersListWithRelationships)
	defer server.Close()

	ctx := context.Background()
	client, err := NewClient(ctx, server.URL, testAPIKey, 3)
	require.NoError(t, err)

	users, nextPageToken, err := client.GetUsers(ctx, "")
	require.NoError(t, err)
	require.Equal(t, "", nextPageToken)
	require.Len(t, users, 3)

	require.Equal(t, "role-owner-guid", users[0].RoleID())
	require.Equal(t, "on-call-role-admin-guid", users[0].OnCallRoleID())

	// a null relationship means the user has no role of that kind
	require.Equal(t, "", users[1].RoleID())
	require.Equal(t, "on-call-role-observer-guid", users[1].OnCallRoleID())

	// a user without a relationships object at all must not panic
	require.Nil(t, users[2].Relationships)
	require.Equal(t, "", users[2].RoleID())
	require.Equal(t, "", users[2].OnCallRoleID())
}
