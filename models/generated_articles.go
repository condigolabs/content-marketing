package models

import "github.com/google/uuid"

type GeneratedArticle struct {
	Runtime       int64    `json:"runtime"`
	Locale        string   `json:"locale"`
	Subject       string   `json:"subject"`
	LabelId       int64    `json:"labelId"`
	Label         string   `json:"label"`
	GeneratedText string   `json:"body"`
	GeneratedHtml string   `json:"html"`
	Images        []string `json:"images"`
	Generator     string   `json:"generator"`
}

func (g *GeneratedArticle) GetTable() string {
	return "generated_articles"
}

func (g *GeneratedArticle) GetDataSet() string {
	return "machinelearning"
}

func (g *GeneratedArticle) Parse() bool {
	return true
}

func (g *GeneratedArticle) GetId() string {
	return uuid.New().String()
}

func (g *GeneratedArticle) SetRunTime(runTime int64) {
	g.Runtime = runTime
}

func (g *GeneratedArticle) IsPartition() bool {
	return false
}
