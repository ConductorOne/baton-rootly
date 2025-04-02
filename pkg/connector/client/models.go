package client

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
	Data  []User `json:"data"`
	Links Links  `json:"links"`
	Meta  Meta   `json:"meta"`
}
