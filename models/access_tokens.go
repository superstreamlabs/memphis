package models

import "time"

type AccessToken struct {
	ID          int       `json:"id"`
	TenantName  string    `json:"tenant_name"`
	IsActive    bool      `json:"is_active"`
	GeneratedBY int       `json:"generated_by"`
	AccessKeyID string    `json:"access_key_id"`
	SecretKey   string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
}

type CreateAccessTokenSchema struct {
	Description string `json:"description"`
}
