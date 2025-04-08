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
	Name               string           `json:"name"`
	Description        string           `json:"description"`
	NotifyEmails       []string         `json:"notify_emails"`
	SlackChannels      []BasicAttribute `json:"slack_channels"`
	SlackAliases       []BasicAttribute `json:"slack_aliases"`
	PagerdutyID        string           `json:"pagerduty_id"`
	PagerdutyServiceID string           `json:"pagerduty_service_id"`
	BackstageID        string           `json:"backstage_id"`
	ExternalID         string           `json:"external_id"`
	OpsGenieID         string           `json:"opsgenie_id"`
	VictorOpsID        string           `json:"victor_ops_id"`
	PagertreeID        string           `json:"pagertree_id"`
	CortexID           string           `json:"cortex_id"`
	ServiceNowCISysID  string           `json:"service_now_ci_sys_id"`
	UserIDs            []string         `json:"user_ids"`
	AdminIDs           []string         `json:"admin_ids"`
	AlertUrgencyID     string           `json:"alert_urgency_id"`
	AlertsEmailEnabled bool             `json:"alerts_email_enabled"`
	AlertsEmailAddress string           `json:"alerts_email_address"`
	UpdatedAt          string           `json:"updated_at"`
	CreatedAt          string           `json:"created_at"`
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
	Data  Team  `json:"data"`
	Links Links `json:"links"`
	Meta  Meta  `json:"meta"`
}
