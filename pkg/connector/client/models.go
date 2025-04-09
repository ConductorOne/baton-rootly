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

type ScheduleRotations struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	// note there's an attributes object available but don't need it for now
}
type ScheduleRotationsResponse struct {
	Data  []ScheduleRotations `json:"data"`
	Links Links               `json:"links"`
	Meta  Meta                `json:"meta"`
}

type ScheduleRotationUserAttributes struct {
	UserID int `json:"user_id"`
	// note there are more attributes available but don't need them for now
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
