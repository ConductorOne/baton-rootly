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
	testPageSize                  = 1
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
        "self": "https://api.example.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "first": "https://api.example.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "prev": null,
        "next": "https://api.example.com/v1/users?page%5Bnumber%5D=2&page%5Bsize%5D=1",
        "last": "https://api.example.com/v1/users?page%5Bnumber%5D=2&page%5Bsize%5D=1"
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
        "self": "https://api.example.com/v1/users?page%5Bnumber%5D=2&page%5Bsize%5D=1",
        "first": "https://api.example.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "prev": "https://api.example.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "next": null,
        "last": "https://api.example.com/v1/users?page%5Bnumber%5D=2&page%5Bsize%5D=1"
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
        "self": "https://api.example.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=2",
        "first": "https://api.example.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=2",
        "prev": null,
        "next": null,
        "last": "https://api.example.com/v1/users?page%5Bnumber%5D=1&page%5Bsize%5D=2"
    },
    "meta": {
        "current_page": 1,
        "next_page": null,
        "prev_page": null,
        "total_count": 2,
        "total_pages": 1
    }
}`
	teamsListResultsPage1of4Size1 = `{
    "data": [
        {
            "id": "sre-team-guid",
            "type": "groups",
            "attributes": {
                "slug": "sre",
                "name": "SRE",
                "description": "The go-to team for all incidents.",
                "color": "#F5D9C4",
                "position": 2,
                "notify_emails": ["test1@example.com", "test2@example.com"],
                "slack_channels": [
                    {
                        "id": "test-channel-id",
                        "name": "test-channel-name"
                    }
                ],
                "slack_aliases": [
                    {
                        "id": "test-alias-id",
                        "name": "test-alias-name"
                    }
                ],
                "pagerduty_id": "test-pagerduty-id",
                "pagerduty_service_id": "test-pagerduty-service-id",
                "backstage_id": "test-backstage-id",
                "external_id": "test-external-id",
                "opsgenie_id": "test-opsgenie-id",
                "victor_ops_id": "test-victor-ops-id",
                "pagertree_id": "test-pagertree-id",
                "cortex_id": "test-cortex-id",
                "service_now_ci_sys_id": "test-service-now-ci-sys-id",
                "user_ids": [
                    96913,
                    97487
                ],
                "admin_ids": [
                    96913
                ],
                "incidents_count": 0,
                "alert_urgency_id": "some-guid",
                "alerts_email_enabled": true,
                "alerts_email_address": "group-test-test@email.rootly.com",
                "created_at": "2025-03-28T07:05:55.007-07:00",
                "updated_at": "2025-04-07T07:54:11.604-07:00"
            }
        }
    ],
    "links": {
        "self": "https://api.example.com/v1/teams?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "first": "https://api.example.com/v1/teams?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "prev": null,
        "next": "https://api.example.com/v1/teams?page%5Bnumber%5D=2&page%5Bsize%5D=1",
        "last": "https://api.example.com/v1/teams?page%5Bnumber%5D=4&page%5Bsize%5D=1"
    },
    "meta": {
        "current_page": 1,
        "next_page": 2,
        "prev_page": null,
        "total_count": 4,
        "total_pages": 4
    }
}`
	teamGetResult = `{
    "data": {
        "id": "sre-team-guid",
        "type": "groups",
        "attributes": {
            "slug": "sre",
            "name": "SRE",
            "description": "The go-to team for all incidents.",
            "color": "#F5D9C4",
            "position": 2,
            "notify_emails": ["test1@example.com", "test2@example.com"],
            "slack_channels": [
                {
                    "id": "test-channel-id",
                    "name": "test-channel-name"
                }
            ],
            "slack_aliases": [
                {
                    "id": "test-alias-id",
                    "name": "test-alias-name"
                }
            ],
            "pagerduty_id": "test-pagerduty-id",
            "pagerduty_service_id": "test-pagerduty-service-id",
            "backstage_id": "test-backstage-id",
            "external_id": "test-external-id",
            "opsgenie_id": "test-opsgenie-id",
            "victor_ops_id": "test-victor-ops-id",
            "pagertree_id": "test-pagertree-id",
            "cortex_id": "test-cortex-id",
            "service_now_ci_sys_id": "test-service-now-ci-sys-id",
            "user_ids": [
                96913,
                97487
            ],
            "admin_ids": [
                96913
            ],
            "incidents_count": 0,
            "alert_urgency_id": "some-guid",
            "alerts_email_enabled": true,
            "alerts_email_address": "group-test-test@email.rootly.com",
            "created_at": "2025-03-28T07:05:55.007-07:00",
            "updated_at": "2025-04-07T07:54:11.604-07:00"
        }
    }
}`
	//nolint:gosec,nolintlint
	secretsListResultsPage1of2Size1 = `{
    "data": [
        {
            "id": "guid-of-my-secret",
            "type": "secrets",
            "attributes": {
                "kind": "built_in",
                "name": "secret_api_token",
                "secret": "[REDACTED]",
                "hashicorp_vault_mount": "secret",
                "hashicorp_vault_path": null,
                "hashicorp_vault_version": 0,
                "created_at": "2025-04-09T06:45:50.486-07:00",
                "updated_at": "2025-04-09T06:45:50.486-07:00"
            }
        }
    ],
    "links": {
        "self": "https://api.example.com/v1/secrets?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "first": "https://api.example.com/v1/secrets?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "prev": null,
        "next": "https://api.example.com/v1/secrets?page%5Bnumber%5D=2&page%5Bsize%5D=1",
        "last": "https://api.example.com/v1/secrets?page%5Bnumber%5D=2&page%5Bsize%5D=1"
    },
    "meta": {
        "current_page": 1,
        "next_page": 2,
        "prev_page": null,
        "total_count": 2,
        "total_pages": 2
    }
}`
	schedulesListResultsPage1of2Size1 = `{
    "data": [
        {
            "id": "test-schedule-guid",
            "type": "schedules",
            "attributes": {
                "name": "Production Oncall",
                "description": null,
                "all_time_coverage": true,
                "slack_user_group": {},
                "owner_user_id": 96913,
                "owner_group_ids": [
                    "sre-team-guid",
                    "security-team-guid"
                ],
                "created_at": "2025-04-07T08:00:54.675-07:00",
                "updated_at": "2025-04-10T16:25:18.032-07:00"
            }
        }
    ],
    "links": {
        "self": "https://api.example.com/v1/schedules?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "first": "https://api.example.com/v1/schedules?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "prev": null,
        "next": "https://api.example.com/v1/schedules?page%5Bnumber%5D=2&page%5Bsize%5D=1",
        "last": "https://api.example.com/v1/schedules?page%5Bnumber%5D=2&page%5Bsize%5D=1"
    },
    "meta": {
        "current_page": 1,
        "next_page": 2,
        "prev_page": null,
        "total_count": 2,
        "total_pages": 2
    }
}`
	scheduleRotationsListResultsPage1of2Size1 = `{
    "data": [
        {
            "id": "test-weekday-rotation-guid",
            "type": "schedule_rotations",
            "attributes": {
                "schedule_id": "test-schedule-guid",
                "name": "rotation weekdays",
                "position": 1,
                "schedule_rotationable_type": "ScheduleWeeklyRotation",
                "active_all_week": false,
                "active_days": [
                    "M",
                    "T",
                    "W",
                    "R",
                    "F"
                ],
                "active_time_type": "all_day",
                "time_zone": "America/Los_Angeles",
                "schedule_rotationable_attributes": {
                    "handoff_time": "00:00",
                    "handoff_day": "M"
                },
                "active_time_attributes": []
            }
        }
    ],
    "links": {
        "self": "https://api.example.com/v1/schedules/test-schedule-guid/schedule_rotations?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "first": "https://api.example.com/v1/schedules/test-schedule-guid/schedule_rotations?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "prev": null,
        "next": "https://api.example.com/v1/schedules/test-schedule-guid/schedule_rotations?page%5Bnumber%5D=2&page%5Bsize%5D=1",
        "last": "https://api.example.com/v1/schedules/test-schedule-guid/schedule_rotations?page%5Bnumber%5D=2&page%5Bsize%5D=1"
    },
    "meta": {
        "current_page": 1,
        "next_page": 2,
        "prev_page": null,
        "total_count": 2,
        "total_pages": 2
    }
}`
	scheduleRotationUsersListResultsPage1of2Size1 = `{
    "data": [
        {
            "id": "test-schedule-rotation-user-guid",
            "type": "schedule_rotation_users",
            "attributes": {
                "schedule_rotation_id": "test-weekday-rotation-guid",
                "user_id": 96913,
                "position": 1,
                "created_at": "2025-04-07T08:00:54.695-07:00",
                "updated_at": "2025-04-07T08:00:54.695-07:00"
            }
        }
    ],
    "links": {
        "self": "https://api.example.com/v1/schedule_rotations/test-weekday-rotation-guid/schedule_rotation_users?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "first": "https://api.example.com/v1/schedule_rotations/test-weekday-rotation-guid/schedule_rotation_users?page%5Bnumber%5D=1&page%5Bsize%5D=1",
        "prev": null,
        "next": "https://api.example.com/v1/schedule_rotations/test-weekday-rotation-guid/schedule_rotation_users?page%5Bnumber%5D=2&page%5Bsize%5D=1",
        "last": "https://api.example.com/v1/schedule_rotations/test-weekday-rotation-guid/schedule_rotation_users?page%5Bnumber%5D=2&page%5Bsize%5D=1"
    },
    "meta": {
        "current_page": 1,
        "next_page": 2,
        "prev_page": null,
        "total_count": 2,
        "total_pages": 2
    }
}`
	scheduleShiftsListResultsPage1of2Size1 = `{
    "data": [
        {
            "id": "test-shift-guid",
            "type": "shifts",
            "attributes": {
                "schedule_id": "test-schedule-guid",
                "rotation_id": null,
                "starts_at": "2025-04-09T12:00:00.000-07:00",
                "ends_at": "2025-04-11T23:59:59.000-07:00",
                "is_override": true
            }
        }
    ],
    "included": [
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
    ]
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
				nextToken:   "https://api.example.com/v1/users?page%5Bnumber%5D=2&page%5Bsize%5D=1",
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

func TestClient_GetTeams(t *testing.T) {
	expectedTeams := []Team{
		{
			ID:   "sre-team-guid",
			Type: "groups",
			Attributes: TeamAttributes{
				Name:        "SRE",
				Description: "The go-to team for all incidents.",
				UserIDs:     []int{96913, 97487},
				AdminIDs:    []int{96913},
				UpdatedAt:   "2025-04-07T07:54:11.604-07:00",
				CreatedAt:   "2025-03-28T07:05:55.007-07:00",
			},
		},
	}
	expectedNextToken := "https://api.example.com/v1/teams?page%5Bnumber%5D=2&page%5Bsize%5D=1" //nolint:gosec,nolintlint
	server := httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set(uhttp.ContentType, "application/json")
				writer.WriteHeader(http.StatusOK)
				_, err := writer.Write([]byte(teamsListResultsPage1of4Size1))
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
		testPageSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	teams, nextPageToken, err := client.GetTeams(ctx, "") // empty page token
	require.Nil(t, err)

	require.Len(t, teams, testPageSize)
	require.Equal(t, expectedTeams[0], teams[0])
	require.ElementsMatch(t, expectedTeams, teams)
	require.Equal(t, expectedNextToken, nextPageToken)
}

func TestClient_GetTeamMemberAndAdminIDs(t *testing.T) {
	teamID := "sre-team-guid"
	expectedMemberIDs := []int{96913, 97487}
	expectedAdminIDs := []int{96913}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				require.Equal(t, "/v1/teams/"+teamID, request.URL.Path)
				writer.Header().Set(uhttp.ContentType, "application/json")
				writer.WriteHeader(http.StatusOK)
				_, err := writer.Write([]byte(teamGetResult))
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
		testPageSize, // doesn't matter for this test
	)
	if err != nil {
		t.Fatal(err)
	}

	memberIDs, adminIDs, err := client.GetTeamMemberAndAdminIDs(ctx, teamID)
	require.Nil(t, err)
	require.ElementsMatch(t, expectedMemberIDs, memberIDs)
	require.ElementsMatch(t, expectedAdminIDs, adminIDs)
}

func TestClient_GetSecrets(t *testing.T) {
	expectedSecrets := []Secret{
		{
			ID:   "guid-of-my-secret",
			Type: "secrets",
			Attributes: SecretAttributes{
				Name:      "secret_api_token",
				UpdatedAt: "2025-04-09T06:45:50.486-07:00",
				CreatedAt: "2025-04-09T06:45:50.486-07:00",
			},
		},
	}
	expectedNextToken := "https://api.example.com/v1/secrets?page%5Bnumber%5D=2&page%5Bsize%5D=1" //nolint:gosec,nolintlint
	server := httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set(uhttp.ContentType, "application/json")
				writer.WriteHeader(http.StatusOK)
				_, err := writer.Write([]byte(secretsListResultsPage1of2Size1))
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
		testPageSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	secrets, nextPageToken, err := client.GetSecrets(ctx, "") // empty page token
	require.Nil(t, err)
	require.Len(t, secrets, testPageSize)
	require.Equal(t, expectedSecrets[0], secrets[0])
	require.ElementsMatch(t, expectedSecrets, secrets)
	require.Equal(t, expectedNextToken, nextPageToken)
}

func TestClient_GetSchedules(t *testing.T) {
	expectedOwnerUserID := 96913
	expectedSchedules := []Schedule{
		{
			ID:   "test-schedule-guid",
			Type: "schedules",
			Attributes: ScheduleAttributes{
				Name:          "Production Oncall",
				Description:   "",
				OwnerUserID:   &expectedOwnerUserID,
				OwnerGroupIDs: []string{"sre-team-guid", "security-team-guid"},
				UpdatedAt:     "2025-04-10T16:25:18.032-07:00",
				CreatedAt:     "2025-04-07T08:00:54.675-07:00",
			},
		},
	}
	expectedNextToken := "https://api.example.com/v1/schedules?page%5Bnumber%5D=2&page%5Bsize%5D=1" //nolint:gosec,nolintlint
	server := httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set(uhttp.ContentType, "application/json")
				writer.WriteHeader(http.StatusOK)
				_, err := writer.Write([]byte(schedulesListResultsPage1of2Size1))
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
		testPageSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	schedules, nextPageToken, err := client.GetSchedules(ctx, "") // empty page token
	require.Nil(t, err)
	require.Len(t, schedules, testPageSize)
	require.Equal(t, expectedSchedules[0], schedules[0])
	require.ElementsMatch(t, expectedSchedules, schedules)
	require.Equal(t, expectedNextToken, nextPageToken)
}

func TestClient_ListScheduleRotations(t *testing.T) {
	testScheduleID := "test-schedule-guid"
	expectedRotationIDs := []string{"test-weekday-rotation-guid"}
	expectedNextToken := "https://api.example.com/v1/schedules/test-schedule-guid/schedule_rotations?page%5Bnumber%5D=2&page%5Bsize%5D=1" //nolint:gosec,nolintlint
	server := httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				require.Equal(t, "/v1/schedules/"+testScheduleID+"/schedule_rotations", request.URL.Path)
				writer.Header().Set(uhttp.ContentType, "application/json")
				writer.WriteHeader(http.StatusOK)
				_, err := writer.Write([]byte(scheduleRotationsListResultsPage1of2Size1))
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
		testPageSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	rotationIDs, nextPageToken, err := client.ListScheduleRotations(ctx, testScheduleID, "") // empty page token
	require.Nil(t, err)
	require.Len(t, rotationIDs, testPageSize)
	require.Equal(t, expectedRotationIDs[0], rotationIDs[0])
	require.ElementsMatch(t, expectedRotationIDs, rotationIDs)
	require.Equal(t, expectedNextToken, nextPageToken)
}

func TestClient_ListScheduleRotationUsers(t *testing.T) {
	testRotationID := "test-weekday-rotation-guid"
	expectedUserIDs := []int{96913}
	expectedNextToken := "https://api.example.com/v1/schedule_rotations/test-weekday-rotation-guid/schedule_rotation_users?page%5Bnumber%5D=2&page%5Bsize%5D=1" //nolint:gosec,nolintlint
	server := httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				require.Equal(t, "/v1/schedule_rotations/"+testRotationID+"/schedule_rotation_users", request.URL.Path)
				writer.Header().Set(uhttp.ContentType, "application/json")
				writer.WriteHeader(http.StatusOK)
				_, err := writer.Write([]byte(scheduleRotationUsersListResultsPage1of2Size1))
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
		testPageSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	userIDs, nextPageToken, err := client.ListScheduleRotationUsers(ctx, testRotationID, "") // empty page token
	require.Nil(t, err)
	require.Len(t, userIDs, testPageSize)
	require.Equal(t, expectedUserIDs[0], userIDs[0])
	require.ElementsMatch(t, expectedUserIDs, userIDs)
	require.Equal(t, expectedNextToken, nextPageToken)
}

func TestClient_ListOnCallUsers(t *testing.T) {
	testScheduleID := "test-schedule-guid"
	expectedUserIDs := []int{97487}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set(uhttp.ContentType, "application/json")
				writer.WriteHeader(http.StatusOK)
				_, err := writer.Write([]byte(scheduleShiftsListResultsPage1of2Size1))
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
		testPageSize, // doesn't matter for this test
	)
	if err != nil {
		t.Fatal(err)
	}

	userIDs, err := client.ListOnCallUsers(ctx, testScheduleID)
	require.Nil(t, err)
	require.ElementsMatch(t, expectedUserIDs, userIDs)
}

func TestClient_generateCurrentPaginatedURL(t *testing.T) {
	ctx := context.Background()
	client, err := NewClient(ctx, testBaseURLStr, testAPIKey, 6)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		pToken         string
		path           string
		pathParameters []string
	}
	tests := []struct {
		name string
		args args
		want *url.URL
	}{
		{
			name: "empty page token, generates url from path and client config values",
			args: args{
				pToken: "",
				path:   "/v1/test",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "v1/test",
				RawQuery: "page%5Bnumber%5D=1&page%5Bsize%5D=6",
			},
		},
		{
			name: "empty page token, generates url with path and path parameters",
			args: args{
				pToken:         "",
				path:           "/v1/tests/%s/%s",
				pathParameters: []string{"guid1", "guid2"},
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "v1/tests/guid1/guid2",
				RawQuery: "page%5Bnumber%5D=1&page%5Bsize%5D=6",
			},
		},
		{
			name: "valid page token, generates url fully from the token",
			args: args{
				pToken: "https://api.example.com/v1/teams?page%5Bnumber%5D=2&page%5Bsize%5D=4",
				path:   "/v1/test",
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "/v1/teams",
				RawQuery: "page%5Bnumber%5D=2&page%5Bsize%5D=4",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := client.generateCurrentPaginatedURL(ctx, tc.args.pToken, tc.args.path, tc.args.pathParameters...)
			// only get an error if the provided path in unparseable, not an interesting test case
			require.Nil(t, err)
			require.Equal(t, tc.want, got)
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
		queryParameters map[string]string
		pathParameters  []string
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
				queryParameters: nil,
			},
			want: &url.URL{
				Scheme: "https",
				Host:   "api.example.com",
			},
		},
		{
			name: "empty query parameters, path starts with backslash",
			args: args{
				path:            "/v1/test",
				queryParameters: map[string]string{},
			},
			want: &url.URL{
				Scheme: "https",
				Host:   "api.example.com",
				Path:   "v1/test",
			},
		},
		{
			name: "no query parameters, path starts with backslash",
			args: args{
				path:            "/v1/test",
				queryParameters: nil,
			},
			want: &url.URL{
				Scheme: "https",
				Host:   "api.example.com",
				Path:   "v1/test",
			},
		},
		{
			name: "no query parameters, one path parameter",
			args: args{
				path:            "/v1/tests/%s",
				queryParameters: nil,
				pathParameters:  []string{"guid"},
			},
			want: &url.URL{
				Scheme: "https",
				Host:   "api.example.com",
				Path:   "v1/tests/guid",
			},
		},
		{
			name: "no query parameters, two path parameters",
			args: args{
				path:            "/v1/tests/%s/%s",
				queryParameters: nil,
				pathParameters:  []string{"guid1", "guid2"},
			},
			want: &url.URL{
				Scheme: "https",
				Host:   "api.example.com",
				Path:   "v1/tests/guid1/guid2",
			},
		},
		{
			name: "no query parameters, use first path parameter",
			args: args{
				path:            "/v1/tests/%s", // only one %s
				queryParameters: nil,
				pathParameters:  []string{"guid1", "guid2"}, // two path parameters
			},
			want: &url.URL{
				Scheme: "https",
				Host:   "api.example.com",
				Path:   "v1/tests/guid1", // only first parameter is used
			},
		},
		{
			name: "no query parameters, path starts without backslash",
			args: args{
				path:            "v1/test",
				queryParameters: nil,
			},
			want: &url.URL{
				Scheme: "https",
				Host:   "api.example.com",
				Path:   "v1/test",
			},
		},
		{
			name: "single query parameter",
			args: args{
				path: "/v1/test",
				queryParameters: map[string]string{
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
			name: "multiple query parameters",
			args: args{
				path: "/v1/test",
				queryParameters: map[string]string{
					"param1": "value1",
					"param2": "123",
					"param3": "true",
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
			name: "multiple query parameters, one path parameter",
			args: args{
				path: "/v1/tests/%s",
				queryParameters: map[string]string{
					"param1": "value1",
					"param2": "123",
					"param3": "true",
				},
				pathParameters: []string{"guid"},
			},
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "v1/tests/guid",
				RawQuery: "param1=value1&param2=123&param3=true",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := client.generateURL(tc.args.path, tc.args.queryParameters, tc.args.pathParameters...)
			require.Equal(t, tc.want, got)
		})
	}
}
