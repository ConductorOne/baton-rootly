package client

import "time"

type User struct {
	ID        string    `jsonapi:"primary,id"`
	Type      string    `jsonapi:"primary,type"`
	Name      string    `jsonapi:"attr,name"`
	Email     string    `jsonapi:"attr,email"`
	FullName  string    `jsonapi:"attr,full_name"`
	SlackID   string    `jsonapi:"attr,slack_id"`
	Phone     string    `jsonapi:"attr,phone"`
	UpdatedAt time.Time `jsonapi:"attr,updated_at"`
	CreatedAt time.Time `jsonapi:"attr,created_at"`
}
