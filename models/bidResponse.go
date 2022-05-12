package models

import "time"

type BidResponse struct {
	RequestID string     `json:"RequestId"`
	Country   string     `json:"Country"`
	Id        string     `json:"Id"`
	LabelId   int64      `json:"labelId"`
	DATE      time.Time  `json:"DATE"`
	Locale    string     `json:"Locale"`
	Label     string     `json:"Label"`
	Brand     string     `json:"Brand"`
	AdChoice  string     `json:"adChoice"`
	Products  []Products `json:"Products"`
}
type Products struct {
	Title       string `json:"Title"`
	Description string `json:"Description"`
	Domain      string `json:"Domain"`
	Brand       string `json:"Brand"`
	Image       string `json:"Image"`
	Target      string `json:"Target"`
}
