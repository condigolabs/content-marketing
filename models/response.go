package models

type SearchResult struct {
	Total      int      `json:"total"`
	TotalPages int      `json:"total_pages"`
	Results    []Result `json:"results"`
}
type Urls struct {
	Raw     string `json:"raw"`
	Full    string `json:"full"`
	Regular string `json:"regular"`
	Small   string `json:"small"`
	Thumb   string `json:"thumb"`
	SmallS3 string `json:"small_s3"`
}
type Links struct {
	Self             string `json:"self"`
	HTML             string `json:"html"`
	Download         string `json:"download"`
	DownloadLocation string `json:"download_location"`
	Photos           string `json:"photos"`
	Likes            string `json:"likes"`
	Portfolio        string `json:"portfolio"`
	Following        string `json:"following"`
	Followers        string `json:"followers"`
}
type ProfileImage struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}
type User struct {
	ID        string `json:"id"`
	UpdatedAt string `json:"updated_at"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
type Tags struct {
	Type  string `json:"type"`
	Title string `json:"title"`
}
type Result struct {
	ID             string `json:"id"`
	Runtime        int64  `json:"runtime"`
	LabelId        int64  `json:"labelId"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	PromotedAt     string `json:"promoted_at"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	Color          string `json:"color"`
	BlurHash       string `json:"blur_hash"`
	Description    string `json:"description"`
	AltDescription string `json:"alt_description"`
	Urls           Urls   `json:"urls"`
	Links          Links  `json:"links"`
	Likes          int    `json:"likes"`
	LikedByUser    bool   `json:"liked_by_user"`
	User           User   `json:"user"`
	Tags           []Tags `json:"tags"`
}

func (g *Result) GetTable() string {
	return "unsplash_images"
}

func (g *Result) GetDataSet() string {
	return "machinelearning"
}

func (g *Result) Parse() bool {
	return true
}

func (g *Result) GetId() string {
	return g.ID
}
func (g *Result) SetRunTime(runTime int64) {
	g.Runtime = runTime
}

func (g *Result) IsPartition() bool {
	return false
}
