package models

type ArticleData struct {
	Title       string `json:"Title"`
	TextContent string `json:"TextContent"`
	Excerpt     string `json:"Excerpt"`
	LabelID     int64  `json:"LabelId"`
	Label       string `json:"Label"`
	Description string `json:"Description"`
	Level       int64
	Urls        struct {
		Raw     string `json:"Raw"`
		Full    string `json:"Full"`
		Regular string `json:"Regular"`
		Small   string `json:"Small"`
		Thumb   string `json:"Thumb"`
		SmallS3 string `json:"SmallS3"`
	} `json:"Urls"`
	Lines []string `json:"lines,omitempty"`
}

func (a ArticleData) GetImage() string {
	return a.Urls.Regular
}
