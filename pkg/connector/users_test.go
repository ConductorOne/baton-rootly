package connector

import (
	"testing"

	"github.com/conductorone/baton-rootly/pkg/connector/client"
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
						CreatedAt: "2023-01-01T00:00:00Z",
						Name:      "Sam",
						FullName:  "Sam Testalot",
						Email:     "sam.testalot@example.com",
						SlackID:   "@testalot",
						Phone:     "123-456-7890",
					},
				},
			},
			want: map[string]interface{}{
				"user_id":    "123",
				"updated_at": "2023-01-01T00:00:00Z",
				"name":       "Sam",
				"full_name":  "Sam Testalot",
				"first_name": "Sam",
				"last_name":  "Testalot",
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
						CreatedAt: "2023-01-02T00:00:00Z",
						Email:     "sam.testalot@example.com",
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
						CreatedAt: "2023-01-03T00:00:00Z",
						FullName:  "Sam Testalot",
						Email:     "sam.testalot@example.com",
					},
				},
			},
			want: map[string]interface{}{
				"user_id":    "125",
				"updated_at": "2023-01-03T00:00:00Z",
				"full_name":  "Sam Testalot",
				"first_name": "Sam",
				"last_name":  "Testalot",
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
