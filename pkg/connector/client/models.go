package client

import (
	"fmt"
	"strings"
)

type Links struct {
	Self  string `json:"self"`
	First string `json:"first"`
	Prev  string `json:"prev"`
	Next  string `json:"next"`
	Last  string `json:"last"`
}

type Meta struct {
	CurrentPage  int `json:"current_page"`
	NextPage     int `json:"next_page"`
	PreviousPage int `json:"prev_page"`
	TotalPages   int `json:"total_pages"`
	TotalCount   int `json:"total_count"`
}

type RootlyError struct {
	Title  string `json:"title"`
	Status string `json:"status"`
	Code   string `json:"code"`   // optional
	Detail string `json:"detail"` // optional
}

// RootlyError represents an error response from the Rootly API.
type RootlyErrorResponse struct {
	Errors []RootlyError `json:"errors"`
}

// Message implements the uhttp.ErrorResponse interface.
func (e *RootlyErrorResponse) Message() string {
	if len(e.Errors) == 0 {
		return "Unknown error from Rootly API"
	}
	var msgs []string
	for _, rootlyError := range e.Errors {
		msg := fmt.Sprintf("%s: %s", rootlyError.Title, rootlyError.Status)
		if rootlyError.Code != "" {
			msg += fmt.Sprintf(", code: %s", rootlyError.Code)
		}
		if rootlyError.Detail != "" {
			msg += fmt.Sprintf(", detail: %s", rootlyError.Detail)
		}
		msgs = append(msgs, msg)
	}
	return strings.Join(msgs, "; ")
}

type UserAttributes struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	FullName  string `json:"full_name"`
	SlackID   string `json:"slack_id"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

// RelationshipData is the JSON:API resource linkage object for a to-one relationship.
type RelationshipData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Relationship is a JSON:API to-one relationship. Data is nil when the relationship is unset.
type Relationship struct {
	Data *RelationshipData `json:"data"`
}

// UserRelationships holds the to-one role relationships Rootly returns alongside each user.
// Rootly models permissions as exactly one Incident Response role and one On-Call role per user.
type UserRelationships struct {
	Role       Relationship `json:"role"`
	OnCallRole Relationship `json:"on_call_role"`
}

type User struct {
	ID            string             `json:"id"`
	Type          string             `json:"type"`
	Attributes    UserAttributes     `json:"attributes"`
	Relationships *UserRelationships `json:"relationships"`
}

// RoleID returns the Incident Response role ID assigned to the user, or "" if there isn't one.
func (u User) RoleID() string {
	if u.Relationships == nil || u.Relationships.Role.Data == nil {
		return ""
	}
	return u.Relationships.Role.Data.ID
}

// OnCallRoleID returns the On-Call role ID assigned to the user, or "" if there isn't one.
func (u User) OnCallRoleID() string {
	if u.Relationships == nil || u.Relationships.OnCallRole.Data == nil {
		return ""
	}
	return u.Relationships.OnCallRole.Data.ID
}

type UsersResponse struct {
	Data  []User `json:"data"`
	Links Links  `json:"links"`
	Meta  Meta   `json:"meta"`
}

type BasicAttribute struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TeamAttributes struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserIDs     []int  `json:"user_ids"`
	AdminIDs    []int  `json:"admin_ids"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type Team struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Attributes TeamAttributes `json:"attributes"`
}

type TeamsResponse struct {
	Data  []Team `json:"data"`
	Links Links  `json:"links"`
	Meta  Meta   `json:"meta"`
}

type TeamResponse struct {
	Data Team `json:"data"`
}

type SecretAttributes struct {
	Name      string `json:"name"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

type Secret struct {
	ID         string           `json:"id"`
	Type       string           `json:"type"`
	Attributes SecretAttributes `json:"attributes"`
}

type SecretsResponse struct {
	Data  []Secret `json:"data"`
	Links Links    `json:"links"`
	Meta  Meta     `json:"meta"`
}

type ScheduleAttributes struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	OwnerUserID   *int     `json:"owner_user_id"`
	OwnerGroupIDs []string `json:"owner_group_ids"`
	UpdatedAt     string   `json:"updated_at"`
	CreatedAt     string   `json:"created_at"`
}

type Schedule struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Attributes ScheduleAttributes `json:"attributes"`
}

type SchedulesResponse struct {
	Data  []Schedule `json:"data"`
	Links Links      `json:"links"`
	Meta  Meta       `json:"meta"`
}

type ScheduleResponse struct {
	Data Schedule `json:"data"`
}

type ScheduleRotationsResponse struct {
	Data  []ObjectWithoutAttributes `json:"data"`
	Links Links                     `json:"links"`
	Meta  Meta                      `json:"meta"`
}

type ScheduleRotationUserAttributes struct {
	UserID int `json:"user_id"`
	// note there are more attributes available but don't need them
}

type ScheduleRotationUser struct {
	ID         string                         `json:"id"`
	Type       string                         `json:"type"`
	Attributes ScheduleRotationUserAttributes `json:"attributes"`
}

type ScheduleRotationUsersResponse struct {
	Data  []ScheduleRotationUser `json:"data"`
	Links Links                  `json:"links"`
	Meta  Meta                   `json:"meta"`
}

type ObjectWithoutAttributes struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	// note there's an attributes object available but don't need or want it
}

type ScheduleShiftsResponse struct {
	// note there's a data object available but don't need it
	Included []ObjectWithoutAttributes `json:"included"`
}

type RoleAttributes struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	IsDeletable bool   `json:"is_deletable"`
	IsEditable  bool   `json:"is_editable"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

// Role is a Rootly Incident Response role, ie a named permission set from /v1/roles.
type Role struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Attributes RoleAttributes `json:"attributes"`
}

type RolesResponse struct {
	Data  []Role `json:"data"`
	Links Links  `json:"links"`
	Meta  Meta   `json:"meta"`
}

type OnCallRoleAttributes struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	// SystemRole is one of admin, user, custom, observer, no_access.
	SystemRole string `json:"system_role"`
	UpdatedAt  string `json:"updated_at"`
	CreatedAt  string `json:"created_at"`
}

// OnCallRole is a Rootly On-Call role, ie a named permission set from /v1/on_call_roles.
type OnCallRole struct {
	ID         string               `json:"id"`
	Type       string               `json:"type"`
	Attributes OnCallRoleAttributes `json:"attributes"`
}

type OnCallRolesResponse struct {
	Data  []OnCallRole `json:"data"`
	Links Links        `json:"links"`
	Meta  Meta         `json:"meta"`
}
