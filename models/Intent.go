package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

type LatestDomain struct {
	Domain string
	Count  int64
}

type LatestProduct struct {
	Date         time.Time `json:"date"`
	RequestId    string    `json:"requestId"`
	Domain       string    `json:"domain"`
	Country      string    `json:"country"`
	Context      string    `json:"context"`
	Product      Product   `json:"product"`
	AvgBid       float64   `json:"avgbid"`
	GenerateLink string    `json:"generateLink"`
}

type ImageIntent struct {
	Label string
	Image string
}
type Intent struct {
	Date    time.Time
	Country string
	Labels  []LabelIntent
}

func (i *Intent) Append(l LabelIntent) {
	i.Labels = append(i.Labels, l)
}

type LabelIntent struct {
	Label        string  `json:"dimensions"`
	Cat          string  `json:"cat"`
	Locale       string  `json:"locale"`
	AvgBid       float64 `json:"avgbid"`
	Count        int64   `json:"count"`
	Score        float64 `json:"score"`
	GenerateLink string  `json:"generateLink"`
}

type Product struct {
	Id          string  `json:"id"`
	Domain      string  `json:"domain"`
	Brand       string  `json:"brand"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Target      string  `json:"target"`
	Image       string  `json:"image"`
	Label       string  `json:"label"`
	Measure     Measure `json:"measures"`
}
type Label struct {
	Label         string
	LabelFullName string
	FriendlyName  string
	Products      []Product
}
type Measure struct {
	AvgBid      float64
	UniqueUsers int64
	Count       int64
}

type GenerateData struct {
	Runtime        int64
	RequestId      string
	LabelId        string
	InputText      []string
	Model          string
	Language       string
	Method         string
	Description    string
	Image          string
	PublisherImage string
}

func (g *GenerateData) GetTable() string {
	return "generated_text"
}

func (g *GenerateData) GetDataSet() string {
	return "machinelearning"
}

func (g *GenerateData) Parse() bool {
	return true
}

func (g *GenerateData) GetId() string {
	return uuid.NewV4().String()
}

func (g *GenerateData) SetRunTime(runTime int64) {
	g.Runtime = runTime
}

func (g *GenerateData) IsPartition() bool {
	return false
}

type Request struct {
	RequestId       string
	PublisherDomain string
	Page            string
	Label           string
	Id              string `json:"id"`
	Domain          string `json:"domain"`
	Brand           string `json:"brand"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Target          string `json:"target"`
	Image           string `json:"image"`
}
