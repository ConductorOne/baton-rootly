package client

import "time"

type Attributes struct {
}

type User struct {
	ID        string    `jsonapi:"primary,id"`
	Name      string    `jsonapi:"attr,name"`
	Email     string    `jsonapi:"attr,email"`
	FullName  string    `jsonapi:"attr,full_name"`
	SlackID   string    `jsonapi:"attr,slack_id"`
	Phone     string    `jsonapi:"attr,phone"`
	UpdatedAt time.Time `jsonapi:"attr,updated_at"`
	CreatedAt time.Time `jsonapi:"attr,created_at"`
}

type Meta struct {
	CurrentPage  int `json:"current_page"`
	NextPage     int `json:"next_page"`
	PreviousPage int `json:"prev_page"`
	TotalPages   int `json:"total_pages"`
	TotalCount   int `json:"total_count"`
}
type UsersResponse struct {
	Data []User `json:"data"`
}
