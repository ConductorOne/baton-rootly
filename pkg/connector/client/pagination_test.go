package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/stretchr/testify/require"
)

// newPaginationTestClient points a client at a stub Rootly that always replies with body.
func newPaginationTestClient(t *testing.T, body string) *Client {
	t.Helper()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set(uhttp.ContentType, "application/json")
				writer.WriteHeader(http.StatusOK)
				_, err := writer.Write([]byte(body))
				if err != nil {
					return
				}
			},
		),
	)
	t.Cleanup(server.Close)

	client, err := NewClient(context.Background(), server.URL, testAPIKey, testPageSize)
	require.NoError(t, err)

	return client
}

// Rootly keeps the links block on the last page and nulls links.next. That is the end of
// the sync, not a failure.
func TestClient_GetUsersLastPage(t *testing.T) {
	client := newPaginationTestClient(t, `{"data":[{"id":"1","type":"users"}],"links":{"self":"s","next":null},"meta":{"current_page":1,"total_pages":1}}`)

	users, nextPageToken, err := client.GetUsers(context.Background(), "")
	require.NoError(t, err)
	require.Len(t, users, 1)
	require.Empty(t, nextPageToken, "the last page should not advance the cursor")
}

// The case the opt-in exists for. A 200 carrying data but no links block reads an empty
// Links.Next, which is indistinguishable from the last page.
func TestClient_GetUsersMissingLinks(t *testing.T) {
	client := newPaginationTestClient(t, `{"data":[{"id":"1","type":"users"}],"meta":{"current_page":1,"total_pages":3}}`)

	_, _, err := client.GetUsers(context.Background(), "")
	require.Error(t, err, "a page without links should not look like the last page")
	require.ErrorIs(t, err, uhttp.ErrMissingPaginationData)
}

func TestClient_GetTeamsMissingLinks(t *testing.T) {
	client := newPaginationTestClient(t, `{"data":[{"id":"1","type":"groups"}],"meta":{"current_page":1,"total_pages":3}}`)

	_, _, err := client.GetTeams(context.Background(), "")
	require.Error(t, err)
	require.ErrorIs(t, err, uhttp.ErrMissingPaginationData)
}

func TestClient_GetSecretsMissingLinks(t *testing.T) {
	client := newPaginationTestClient(t, `{"data":[{"id":"1","type":"secrets"}],"meta":{"current_page":1,"total_pages":3}}`)

	_, _, err := client.GetSecrets(context.Background(), "")
	require.Error(t, err)
	require.ErrorIs(t, err, uhttp.ErrMissingPaginationData)
}

// TeamResponse is a single-resource read with no pagination envelope, so it must not start
// demanding one even though it shares the same request path.
func TestClient_GetTeamMemberAndAdminIDsDoesNotRequirePagination(t *testing.T) {
	client := newPaginationTestClient(t, `{"data":{"id":"t1","type":"groups","attributes":{"name":"SRE","user_ids":[1,2],"admin_ids":[1]}}}`)

	memberIDs, adminIDs, err := client.GetTeamMemberAndAdminIDs(context.Background(), "t1")
	require.NoError(t, err)
	require.ElementsMatch(t, []int{1, 2}, memberIDs)
	require.ElementsMatch(t, []int{1}, adminIDs)
}
