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

// paginationEnvelope is the JSON:API envelope Rootly returns on every collection
// endpoint. Links is a pointer so a response that arrived without the block is
// distinguishable from the last page, where Rootly keeps the object and nulls links.next.
type paginationEnvelope struct {
	Links *Links `json:"links"`
	Meta  Meta   `json:"meta"`
}

// HasPaginationData satisfies uhttp.PaginatedResponse, so a page that came back without
// its links block fails the request instead of reading an empty Links.Next and silently
// ending the sync.
func (p *paginationEnvelope) HasPaginationData() bool {
	return p.Links != nil
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

type User struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Attributes UserAttributes `json:"attributes"`
}

type UsersResponse struct {
	Data []User `json:"data"`
	paginationEnvelope
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
	Data []Team `json:"data"`
	paginationEnvelope
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
	Data []Secret `json:"data"`
	paginationEnvelope
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
	Data []Schedule `json:"data"`
	paginationEnvelope
}

type ScheduleResponse struct {
	Data Schedule `json:"data"`
}

type ScheduleRotationsResponse struct {
	Data []ObjectWithoutAttributes `json:"data"`
	paginationEnvelope
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
	Data []ScheduleRotationUser `json:"data"`
	paginationEnvelope
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
