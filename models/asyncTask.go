package models

import "time"

type AsyncTask struct {
	ID            int         `json:"id"`
	Name          string      `json:"name"`
	BrokrInCharge string      `json:"broker_in_charge"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	Data          interface{} `json:"data"`
	TenantName    string      `json:"tenantName"`
}

type MetaData struct {
	Offset int `json:"offset"`
}
