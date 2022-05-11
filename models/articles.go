package models

import (
	"time"
)

type RawData struct {
	Uri     string
	LabelId int64
	FullId  string
	Label   string
	Locale  string
}
type Article struct {
	Date        time.Time `json:"data,omitempty"`
	Runtime     int64     `json:"-,omitempty"`
	Domain      string    `json:"domain,omitempty"`
	Uri         string    `json:"uri,omitempty"`
	Status      int64     `json:"status,omitempty"`
	Title       string    `json:"title,omitempty"`
	Byline      string    `json:"byline,omitempty"`
	Content     string    `json:"-,omitempty"`
	TextContent string    `json:"-,omitempty"`
	Length      int       `json:"length,omitempty"`
	Excerpt     string    `json:"excerpt,omitempty"`
	SiteName    string    `json:"site_name,omitempty"`
	Image       string    `json:"image,omitempty"`
	Favicon     string    `json:"favicon,omitempty"`
	LabelId     int64     `json:"label_id,omitempty"`
	Label       string    `json:"label,omitempty"`
	Lines       []string  `json:"lines,omitempty"`
}

func (g *Article) GetTable() string {
	return "readable_article"
}

func (g *Article) GetDataSet() string {
	return "machinelearning"
}

func (g *Article) Parse() bool {
	return true
}

func (g *Article) GetId() string {
	return g.Uri
}

func (g *Article) SetRunTime(runTime int64) {
	g.Runtime = runTime
}

func (g *Article) IsPartition() bool {
	return false
}
