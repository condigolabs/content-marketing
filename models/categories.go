package models

type Categories struct {
	Runtime      int64  `json:"runtime"`
	Label        string `json:"label"`
	ID           int64  `json:"id"`
	FriendlyName string `json:"friendlyName"`
	Count        int64  `json:"count"`
}
