package models

type DeleteUserRequest struct {
	Usernames  []string `json:"users"`
	TenantName string   `json:"tenant_name"`
}
