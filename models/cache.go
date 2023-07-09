package models

type CacheUpdateRequest struct {
	CacheType  string   `json:"type"`
	Operation  string   `json:"operation"`
	Usernames  []string `json:"users"`
	TenantName string   `json:"tenant_name"`
}
