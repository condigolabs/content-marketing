package controller

import (
	"context"
	"fmt"
	"github.com/condigolabs/content-marketing/models"
	"github.com/condigolabs/content-marketing/services/intent"
	"github.com/condigolabs/content-marketing/services/reader"
	"github.com/condigolabs/content-marketing/startup"
	domainmodels "github.com/condigolabs/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var root = "https://content-marketing.cdglb.com"

func InitRouter(engine *gin.Engine) {
	intents := engine.Group("/intents")
	intents.GET("bids", GetLatestProducts)
	intents.GET("intents", GetLatestIntent)
	intents.GET("request/:id/:ext", GetArticlesFromRequest)
	intents.GET("cat/:tag/:ext", GetFromTag)
	intents.GET("view", Read)
	intents.GET("parseall", Parse)
	intents.GET("lookup", LookupArticles)
	intents.GET("generate", GenerateArticles1)
	intents.GET("test", TestTemplate)
	intents.GET("bidresponse", LoadBidRequest)
	intents.GET("categories", LoadCategories)
}

type Dimensions struct {
	Type   string   `json:"type"`
	Label  string   `json:"label"`
	Values []string `json:"values"`
}
type Tag struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
	Link  string `json:"link"`
}
type LatestProducts struct {
	Dimensions []Dimensions           `json:"dimensions"`
	Bids       []models.LatestProduct `json:"bids"`
	Tags       []Tag                  `json:"tags"`
}

func GetLatestProducts(c *gin.Context) {
	service := startup.GetIntent()
	param := intent.Param{
		Country:   "USA",
		LastHours: 120,
		Locale:    "en-US",
	}
	ret, err := service.LoadLatestProducts(param)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}

	for i := 0; i < len(ret); i++ {
		u, _ := url.Parse(fmt.Sprintf("%s/intents/generate", root))

		q := u.Query()
		q.Add("t", ret[i].Product.Label)
		q.Add("lid", fmt.Sprintf("%d", ret[i].LabelId))
		q.Add("w", "true")
		u.RawQuery = q.Encode()
		ret[i].GenerateLink = u.String() //fmt.Sprintf("%s/intents/request/%s/html", root, ret[i].RequestId)
	}
	tags, err := service.LoadLatestIntent(param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	vTags := make([]Tag, len(tags.Labels))
	for i := 0; i < len(vTags); i++ {
		u, _ := url.Parse(fmt.Sprintf("%s/intents/generate", root))

		q := u.Query()
		q.Add("t", tags.Labels[i].Label)
		q.Add("lid", fmt.Sprintf("%d", tags.Labels[i].ID))
		q.Add("w", "true")
		u.RawQuery = q.Encode()

		vTags[i] = Tag{
			Value: tags.Labels[i].Label,
			Count: tags.Labels[i].Count,
			Link:  u.String(),
		}
	}

	c.JSON(http.StatusOK, LatestProducts{
		Tags:       vTags,
		Dimensions: []Dimensions{},
		Bids:       ret,
	})
}

func GetLatestIntent(c *gin.Context) {
	service := startup.GetIntent()
	ret, err := service.LoadLatestIntent(intent.Param{
		Country:   "USA",
		LastHours: 120,
		Locale:    "en-US",
	})

	for i := 0; i < len(ret.Labels); i++ {
		ret.Labels[i].GenerateLink = fmt.Sprintf("https://content-marketing.cdglb.com/intents/cat/%s/html", ret.Labels[i].Cat)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	c.JSON(http.StatusOK, ret)
}

type Generated struct {
	GeneratedText []models.GenerateData
	Images        []string
}

func GetRequest(c *gin.Context) {

	service := startup.GetIntent()

	requestId := c.Param("id")
	method := c.Param("ext")

	var ret []models.GenerateData

	if len(requestId) > 0 {
		request, err := service.LoadRequest(intent.Param{
			Locale:    "en-US",
			RequestId: c.Param("id"),
		})
		if err == nil {

			g, err := service.LoadGenerated(c.Param("id"))
			if err == nil && len(g) > 0 {
				ret = g
				if method == "html" {
					c.HTML(http.StatusOK, "default.tmpl.html", ret)
				} else {
					c.JSON(http.StatusOK, ret)
				}

				return
			}

			for _, i := range request {
				headline := intent.DocHeadlines{}
				headline.Labels = append(headline.Labels, i.Title)
				if len(i.Label) > 0 {
					headline.Labels = append(headline.Labels, i.Label)
				}
				if len(i.Brand) > 0 {
					headline.Labels = append(headline.Labels, i.Brand)
				}
				d, err := service.GenerateDocument(headline)
				if err == nil {
					d.RequestId = requestId
					d.Image = i.Image
					ret = append(ret, d)
					err = service.FlushEntitySync(&d)
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
	if method == "html" {
		c.HTML(http.StatusOK, "default.tmpl.html", ret)
	} else {
		c.JSON(http.StatusOK, ret)
	}
}

func GetFromTag(c *gin.Context) {

	service := startup.GetIntent()

	tag := c.Param("tag")
	method := c.Param("ext")
	var ret []models.GenerateData

	if len(tag) > 0 {

		img, err := service.LoadRandImage(intent.Param{
			Tag: tag,
		})
		if err == nil {
			for _, i := range img {
				headline := intent.DocHeadlines{}
				headline.Labels = append(headline.Labels, tag)
				headline.Labels = append(headline.Labels, i.Label)
				d, err := service.GenerateDocument(headline)
				if err == nil {
					d.Image = i.Image
					ret = append(ret, d)
					err = service.FlushEntitySync(&d)
				}
			}
		}
	}
	if method == "html" {
		c.HTML(http.StatusOK, "default.tmpl.html", ret)
	} else {
		c.JSON(http.StatusOK, ret)
	}
}

type ReadParam struct {
	Uri        string `form:"uri"`
	Type       string `form:"type"`
	QueryTerms string `form:"q"`
	Wide       bool   `form:"w"`
	Title      string `form:"t"`
	Label      string `form:"c"`
	LabelId    int64  `form:"lid"`
}

func Parse(c *gin.Context) {

	service := startup.GetIntent()
	out := make(chan models.RawData)

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		defer close(out)
		return service.LoadRawData(ctx,
			intent.Param{
				Country:   "USA",
				LastHours: 120,
				Locale:    "en-US",
			}, out)

	})
	for i := 0; i < 16; i++ {
		for r := range out {
			a, err := reader.Do(r.Uri, r.LabelId)
			if err == nil {
				err = service.FlushEntitySync(&a)
				if err != nil {
					logrus.WithError(err).Errorf("ERRRO")
				}
			}
		}
	}
	if err := g.Wait(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
	return
}
func Read(c *gin.Context) {
	var m ReadParam

	c.ShouldBindQuery(&m)

	a, err := reader.Do(m.Uri, 0)
	if err != nil {

	}
	if m.Type == "text" {
		c.Writer.WriteString(a.TextContent)
	} else {
		c.Writer.WriteString(a.Content)
	}
}

func LookupArticles(c *gin.Context) {
	service := startup.GetIntent()
	var m ReadParam
	p := intent.LookupArticleParam{

		QueryTerms: m.QueryTerms,
	}
	c.ShouldBindQuery(&m)
	if m.Wide {
		p.Fields = "Label,Title,TextContent"
	} else {
		p.Fields = "Label,Title"
	}

	ret, err := service.LookupArticlesNew(p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, ret)
}
func generate(m ReadParam) ([]models.Article, error) {
	service := startup.GetIntent()
	p := intent.LookupArticleParam{
		QueryTerms: m.QueryTerms,
	}
	if m.Wide {
		p.Fields = "Label,Title,TextContent"
	} else {
		p.Fields = "Label,Title"
	}
	if len(p.QueryTerms) == 0 {
		p.QueryTerms = fmt.Sprintf("%s,%s", m.Title, m.Label)
	}
	return service.LookupArticles(p)
}
func ToHtml(c *gin.Context, ret []models.Article) {
	rand.Seed(time.Now().UnixNano())
	payload := PayloadArticleContent{
		PublisherName:      "IlliPress",
		PublisherLogo:      "https://s3.us-east-1.amazonaws.com/cdglb-content-server/img/sites/J2CPaXFfi/5gEiWdJKZ_full.jpg",
		ArticleTitle:       ret[0].Title,
		ArticleDescription: ret[0].Excerpt,
		Author:             "Magic Condigolabs ",
		MainImage:          ret[0].Image,
		Quote: Item{
			Title:    PickQuotes(),
			HeadLine: PickQuotes(),
		},
	}

	for _, l := range strings.Split(ret[0].Label, ">") {

		payload.BreadCrumb = append(payload.BreadCrumb, Action{
			Name: l,
			Link: "#",
		})
	}
	for idx, a := range ret {
		i := Item{
			Id:       fmt.Sprintf("placement-%d", idx),
			Title:    a.Title,
			Image:    a.Image,
			HeadLine: a.Excerpt,
		}

		for idx, l := range a.Lines {
			format := "header"
			if idx > 0 {
				format = "paragraph"
				if rand.Float64() < 0.2 {
					format = "header"
				}
			}
			i.Lines = append(i.Lines, Line{
				Text:   l,
				Format: format,
			})

		}
		payload.Articles = append(payload.Articles, i)
	}
	c.HTML(http.StatusOK, "index.tmpl.html", payload)
}
func GenerateArticles1(c *gin.Context) {
	service := startup.GetIntent()
	var m ReadParam
	c.ShouldBindQuery(&m)

	p := intent.LookupArticleParam{
		QueryTerms: m.QueryTerms,
		LabelId:    m.LabelId,
	}

	if m.Wide {
		p.Fields = "Label,Title,TextContent"
	} else {
		p.Fields = "Label,Title"
	}
	if len(p.QueryTerms) == 0 {
		p.QueryTerms = fmt.Sprintf("%s,%s", m.Title, m.Label)
	}
	ret, err := service.LookupArticlesNew(p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if len(ret) > 0 {
		payload := PayloadArticleContent{
			PublisherName:      "IlliPress",
			PublisherLogo:      "https://s3.us-east-1.amazonaws.com/cdglb-content-server/img/sites/J2CPaXFfi/5gEiWdJKZ_full.jpg",
			ArticleTitle:       ret[0].Title,
			ArticleDescription: ret[0].Excerpt,
			Author:             "Magic Condigolabs ",
			MainImage:          ret[0].GetImage(),
			Quote: Item{
				Title:    PickQuotes(),
				HeadLine: PickQuotes(),
			},
		}

		for _, l := range strings.Split(ret[0].Label, ">") {

			payload.BreadCrumb = append(payload.BreadCrumb, Action{
				Name: l,
				Link: "#",
			})
		}
		for idx, a := range ret {
			i := Item{
				Id:       fmt.Sprintf("placement-%d", idx),
				Title:    a.Title,
				Image:    a.GetImage(),
				HeadLine: a.Excerpt,
			}

			bids, err := service.LoadBidRequest(intent.Param{
				LabelId: a.LabelID,
				Locale:  "en-US",
			})
			if err == nil && len(bids) > 0 {

				i.Ads = bids[0]
				i.Ads.Id = fmt.Sprintf("placement-%d", idx)
				i.Ads.Brand = bids[0].Products[0].Brand
				i.Ads.AdChoice = "https://privacy.eu.criteo.com/adchoice"
			}

			for idx, l := range a.Lines {
				format := "header"
				if idx > 0 {
					format = "paragraph"
					if rand.Float64() < 0.2 {
						format = "header"
					}
				}
				i.Lines = append(i.Lines, Line{
					Text:   l,
					Format: format,
				})

			}
			payload.Articles = append(payload.Articles, i)
		}
		c.HTML(http.StatusOK, "index.tmpl.html", payload)
		return
	}
	TestTemplate(c)

}
func GenerateArticlesX(c *gin.Context) {
	service := startup.GetIntent()
	var m ReadParam
	p := intent.LookupArticleParam{
		QueryTerms: m.QueryTerms,
	}
	c.ShouldBindQuery(&m)
	if m.Wide {
		p.Fields = "Label,Title,TextContent"
	} else {
		p.Fields = "Label,Title"
	}
	if len(p.QueryTerms) == 0 {
		p.QueryTerms = fmt.Sprintf("%s,%s", m.Title, m.Label)
	}
	ret, err := service.LookupArticles(p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if len(ret) > 0 {
		payload := PayloadArticleContent{
			PublisherName:      "IlliPress",
			PublisherLogo:      "https://s3.us-east-1.amazonaws.com/cdglb-content-server/img/sites/J2CPaXFfi/5gEiWdJKZ_full.jpg",
			ArticleTitle:       ret[0].Title,
			ArticleDescription: ret[0].Excerpt,
			Author:             "Magic Condigolabs ",
			MainImage:          ret[0].Image,
			Quote: Item{
				Title:    PickQuotes(),
				HeadLine: PickQuotes(),
			},
		}

		for _, l := range strings.Split(ret[0].Label, ">") {

			payload.BreadCrumb = append(payload.BreadCrumb, Action{
				Name: l,
				Link: "#",
			})
		}
		for idx, a := range ret {
			i := Item{
				Id:       fmt.Sprintf("placement-%d", idx),
				Title:    a.Title,
				Image:    a.Image,
				HeadLine: a.Excerpt,
				Lines:    nil,
				Actions:  nil,
				Ads:      models.BidResponse{},
			}

			for idx, l := range a.Lines {
				format := "header"
				if idx > 0 {
					format = "paragraph"
					if rand.Float64() < 0.2 {
						format = "header"
					}
				}
				i.Lines = append(i.Lines, Line{
					Text:   l,
					Format: format,
				})

			}
			payload.Articles = append(payload.Articles, i)
		}
		c.HTML(http.StatusOK, "index.tmpl.html", payload)
		return
	}
	TestTemplate(c)
	/*input, err := service.ApplyTemplate(intent.PayloadTemplate{
		InputTitle: m.Title,
		InputLabel: m.Label,
		Articles:   ret,
	})

	c.Writer.WriteString(input)*/

}
func GetArticlesFromRequest(c *gin.Context) {

	service := startup.GetIntent()
	requestId := c.Param("id")

	if len(requestId) > 0 {
		request, err := service.LoadRequest(intent.Param{
			Locale:    "en-US",
			RequestId: c.Param("id"),
		})
		if err == nil {
			var all []models.Article
			for _, i := range request {
				ret, err := generate(ReadParam{
					QueryTerms: fmt.Sprintf("%s %s %s", i.Title, i.Label, i.Brand),
					Wide:       true,
				})
				if err == nil {
					all = append(all, ret...)
				}
			}
			if len(all) > 0 {
				ToHtml(c, all)
				return
			}
		}
	}
	TestTemplate(c)

}

type Action struct {
	Name string
	Link string
}
type Line struct {
	Text   string
	Format string
}
type Item struct {
	Id       string
	Title    string
	Image    string
	HeadLine string
	Lines    []Line
	Actions  []Action
	Ads      models.BidResponse
}

type Placement struct {
	Id   string
	Data []domainmodels.ResponseCategories
}
type PayloadArticleContent struct {
	PublisherName      string
	PublisherLogo      string
	ArticleTitle       string
	ArticleDescription string
	MainImage          string
	Author             string
	BreadCrumb         []Action
	Articles           []Item
	Quote              Item
	Placement          Placement
}

var images = []string{
	"https://images.unsplash.com/photo-1580715201266-c1474931ff26?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1526385159909-196a9ac0ef64?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1518057111178-44a106bad636?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1558160074-4d7d8bdf4256?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1554427518-efad17b1dfae?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1513694203232-719a280e022f?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1536008046477-01746710ffb9?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1565708305829-9aeef92a6262?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1504194008492-c55ffe34e18d?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1580147045522-d62c309daa4a?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1580529448475-8ebf21c90d0a?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1455582916367-25f75bfc6710?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1534162802244-d6f69e9048da?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1483794344563-d27a8d18014e?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1459664018906-085c36f472af?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1502977249166-824b3a8a4d6d?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1487139975590-b4f1dce9b035?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1474557157379-8aa74a6ef541?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1502230831726-fe5549140034?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1524901548305-08eeddc35080?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1580465446361-8aae5321522b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1530522238647-b1b7e1789c39?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1494790108377-be9c29b29330?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1489710437720-ebb67ec84dd2?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1479936343636-73cdc5aae0c3?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1525206809752-65312b959c88?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1562337404-3044c84ac061?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1533227268428-f9ed0900fb3b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1489278353717-f64c6ee8a4d2?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1568684333892-fdce35bd742b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1575444095834-a897c7343132?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1495420378468-78588a508652?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1518417600321-552a796eb55e?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1519307335631-ab6963fda5bf?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1511679748904-1d4b70e91b1b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1563373983-2eb50f3226f8?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1471958680802-1345a694ba6d?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1545764964-094fd72bd8d2?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1517428956741-0e738679fc79?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1542794961-5647fd97e6f3?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1547803437-56a009e90946?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1522865389096-9e6e525333d4?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1518406002662-edf7ea6c3b41?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1529314317205-42e5009e8f08?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1518531933037-91b2f5f229cc?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1492496913980-501348b61469?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1521305916504-4a1121188589?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1576618148367-557c39975095?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1563612116891-9b03e4bb9318?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1526318896980-cf78c088247c?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1565299585323-38d6b0865b47?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1568051243851-f9b136146e97?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1545288907-ffa8bdb07477?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1541971897566-308cf7ad0934?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1510812431401-41d2bd2722f3?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1531947398206-60f8e97f34a2?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1515150144380-bca9f1650ed9?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1561451407-3f768c6133ce?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1534816788524-f89604e58abb?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1517944512237-30c5de26233d?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1545746095-f2dd7386466b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1551712606-8d1a1ff352e9?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1536147116438-62679a5e01f2?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1505816014357-96b5ff457e9a?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1424772684780-a05a720ff374?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1500081340404-7f608b644c78?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1468817814611-b7edf94b5d60?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1513094735237-8f2714d57c13?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1512794268250-65fd4cd7441f?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1514580426463-fd77dc4d0672?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1506667527953-22eca67dd919?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1528826792843-696806ff0b87?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1483985988355-763728e1935b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1485579149621-3123dd979885?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1539903259928-7aa0280ce0ad?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1540825002004-f3f9d60e388f?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1564509845570-bd602d5ed6f5?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1564509845953-09b96471b7cb?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1564518025118-46897fe4b0e7?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1564850478228-a582ed113cd8?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1505682634904-d7c8d95cdc50?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1487180144351-b8472da7d491?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1524117304818-b4fadd3e127a?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1546435770-a3e426bf472b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1500627965408-b5f2c8793f17?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1474354503580-955e733d2a7d?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1526948128573-703ee1aeb6fa?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1485827404703-89b55fcc595e?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1527443060795-0402a18106c2?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1491947153227-33d59da6c448?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1524504211093-49fc246db7ed?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1487058792275-0ad4aaf24ca7?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1547394765-185e1e68f34e?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1507120410856-1f35574c3b45?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1544367567-0f2fcb009e0b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1509833903111-9cb142f644e4?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1445384763658-0400939829cd?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1518459031867-a89b944bffe4?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1470468969717-61d5d54fd036?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1538805060514-97d9cc17730c?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1551468307-8c1e3c78013c?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1474888505161-1ace11ae3d81?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1467632499275-7a693a761056?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1470101691117-2571c356a668?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1540845692348-b9d2bc813a63?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1498842812179-c81beecf902c?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1531853121101-cb94c8ed218d?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1489844097929-c8d5b91c456e?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1542435503-956c469947f6?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1568658175033-c815818d2c1b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1491438590914-bc09fcaaf77a?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1543269664-56d93c1b41a6?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1483389127117-b6a2102724ae?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1517245386807-bb43f82c33c4?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1529400971008-f566de0e6dfc?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1548203000-9d0ebf197095?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1575176218117-dfdc2fa48902?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1524502397800-2eeaad7c3fe5?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1551184451-76b762941ad6?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1529408570047-e4414fb17e95?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1516063131707-07d5952d90cc?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1529432337323-223e988a90fb?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1534515729281-5ddf2c470538?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1523371542221-b2965445c8f3?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1532527129402-82f924cb0ed1?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1474325874720-4b395be870c4?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1529672425113-d3035c7f4837?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1479752524501-2a1efb81c407?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1476234251651-f353703a034d?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1453342664588-b702c83fc822?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1571771826307-98d0d0999028?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1560707854-fb9a10eeaace?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1559561875-dbbdf3700044?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1525956180549-4d511fd16335?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1531983412531-1f49a365ffed?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1516641239768-dc3572bdca04?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1511300542434-16b61e1ce871?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1537735319956-df7db4b6a4e9?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1514415008039-efa173293080?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1566616213894-2d4e1baee5d8?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1554062614-6da4fa67725a?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1508721713313-60b1109e2d4b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1499756630622-6a7fd76720ab?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1537476102677-80bac0ab1d8b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1531310197839-ccf54634509e?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1515347619252-60a4bf4fff4f?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1515224526905-51c7d77c7bb8?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1500917832468-298fa6292e2b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1580383538415-f60a57683a69?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1551010786-14eca273a8fc?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1530521954074-e64f6810b32d?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1470596914251-afb0b4510279?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1573935417998-6a3f9bdefb50?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1529171308272-201c63f8831a?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1580301579799-aac8aad4e96b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1505761671935-60b3a7427bad?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1542691246-88d7c605ce87?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1502736885509-af72c92db31a?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1514944666244-c656564a01fa?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1520523839897-bd0b52f945a0?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1513883049090-d0b7439799bf?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1514612497953-05d1e5e171fa?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1558882268-15aa056d885f?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1535732820275-9ffd998cac22?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1507679252487-e3db58b1642e?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1490367532201-b9bc1dc483f6?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1543791187-df796fa11835?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1467647160393-708009aefd5c?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1514891163508-4d0d04535922?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1476611317561-60117649dd94?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1501426026826-31c667bdf23d?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1486403184395-fc4990866136?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1518609878373-06d740f60d8b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1553830591-d8632a99e6ff?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1519020925855-604f4b23d490?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1471275382285-68e99fabd60a?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1554257281-3dba342159d5?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1461468611824-46457c0e11fd?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1504203772830-87fba72385ee?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1527610276295-f4c1b38decc5?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1567540017993-c888313000b1?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1489549132488-d00b7eee80f1?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1532361825603-f5a541a09cc4?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1555445091-5a8b655e8a4a?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1555445122-bc2ba06500be?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1580331522941-58d0812c6d81?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1520322082799-20c1288346e3?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1490049350474-498de43bc885?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1529854140025-25995121f16f?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1531379410502-63bfe8cdaf6f?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1533258439784-28006397342d?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1557182307-a2fd5bcafedd?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1519710164239-da123dc03ef4?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1570169043013-de63774bbf97?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1477414348463-c0eb7f1359b6?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1508873787497-1b513a18217a?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1546098073-e1df70b3bc7c?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1507039929664-6ccc1a9d36dc?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1488485282435-e2ad51917a76?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1505904267569-f02eaeb45a4c?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1429087969512-1e85aab2683d?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1520512533001-af75c194690b?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1455541029258-597a69778eed?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1529333166437-7750a6dd5a70?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1491438590914-bc09fcaaf77a?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1511988617509-a57c8a288659?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
	"https://images.unsplash.com/photo-1511632765486-a01980e01a18?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=400&fit=max&ixid=eyJhcHBfaWQiOjExMzkxNH0",
}
var quotes = []string{
	"If you can dream, you can do it ..",
	"The happiest people do not have everything best. They just do their best with everything they have",
	"With too much we get lost. With less we are",
	"Do not seek happiness, create it",
	"Better than perfect",
	"Believe in your dreams and they may come true. Believe in you and they will surely come true",
	"The heart knows the right answer.",
}

type T struct {
	Lines []Line
}

var Articles = []T{
	{
		Lines: []Line{
			{
				Text:   "including a variety of pot and planter ideas for your garden!",
				Format: "header",
			},
			{Text: "When it comes to gardening, pots and planters are a gardener’s best friend.",
				Format: "paragraph"},
			{Text: "They can add color, life, and personality to any garden, and can be used to grow a variety of plants, flowers, and vegetables.",
				Format: "paragraph"},
			{Text: "There are a variety of different pots and planters that you can use in your garden, and the options are endless. Here are a few ideas to get you started:",
				Format: "header"},
			{Text: ".1. Clay pots: Clay pots are a classic choice for pots and planters. They come in a variety of shapes and sizes, and are a great option for both indoor and outdoor gardens..",
				Format: "paragraph"},
			{Text: "2. Ceramic pots: Ceramic pots are another popular choice for pots and planters. They come in a variety of colors and styles, and are a great option for both indoor and outdoor gardens..",
				Format: "paragraph"},
			{Text: "3. Plastic pots: Plastic pots are a budget-friendly option, and are a great choice for both indoor and outdoor gardens. They come in a variety of shapes and sizes, and are a great option for both plants and flowers..",
				Format: "paragraph"},
			{Text: "4. Wooden planters: Wooden planters are a beautiful option for outdoor gardens, and can be used to grow a variety of plants and flowers..", Format: "paragraph"},
			{Text: "5. Metal planters: Metal planters are a great option for both indoor and outdoor gardens, and come in a variety of shapes and sizes..", Format: "paragraph"},
			{Text: "6. Stone pots: Stone pots are a beautiful option for outdoor gardens, and can be used to grow a variety of plants and flowers..", Format: "paragraph"},
			{Text: "7. Hanging pots: Hanging pots are a great option for both indoor and outdoor gardens, and are perfect for growing plants and flowers..", Format: "paragraph"},
			{Text: "8. Window boxes: Window boxes are a great option for both indoor and outdoor gardens, and are perfect for growing plants and flowers..", Format: "paragraph"},
			{Text: "9. Self-watering pots: Self-watering pots are a great option for both indoor and outdoor gardens, and are perfect for growing plants that need a lot of water.", Format: "paragraph"},
			{Text: "10. And finally don’t forget to add some personality to your pots and planters with a little bit of paint or some fun accessories!", Format: "paragraph"},
		},
	},
	{
		Lines: []Line{
			{Text: "Pots & Planters", Format: "header"},
			{Text: "What could be more cheerful than a pot of brightly blooming flowers on your porch or patio? Flower pots come in all shapes, sizes, and colors, and can be used to add a splash of color to any outdoor space.", Format: "paragraph"},
			{Text: "Choose pots that are large enough to accommodate the plants you want to grow, and make sure the pot has drainage holes so the water can escape. You may also want to consider adding a potting mix to the soil to improve drainage and help the soil retain moisture.", Format: "paragraph"},
			{Text: "When selecting plants for your pots and planters, be sure to choose ones that are suited for your climate and growing conditions. If you live in a cold climate, choose plants that can tolerate frosty weather, and if you live in a warmer climate, choose plants that can tolerate hot weather.", Format: "paragraph"},
			{Text: "Hundreds of beautiful flowers and plants are available for planting in pots and planters, so you're sure to find ones that will suit your taste and your climate.Here are a few of our favorites:", Format: "paragraph"},
			{Text: "Petunias are a popular choice for flower pots, and come in a variety of colors including pink, red, purple, and white. They thrive in sunny locations and can tolerate some cold weather.", Format: "header"},
			{Text: "Lantanas are another popular choice, and come in a variety of colors including yellow, red, orange, and pink. They also thrive in sunny locations, and can tolerate some cold weather.", Format: "paragraph"},
			{Text: "Begonias are attractive plants that come in a variety of colors including pink, red, white, and orange. They do well in shady locations, and can tolerate some cold weather.", Format: "paragraph"},
			{Text: "Impatiens are brightly colored plants that come in a variety of colors including pink, red, purple, and white. They do well in shady locations, and can tolerate some cold weather.", Format: "header"},
			{Text: "When planting flowers in pots and planters, be sure to keep the plants well watered, especially during hot weather. You may also want to fertilize the plants every few weeks.", Format: "paragraph"},
			{Text: "A pot of brightly blooming flowers can add color and charm to any outdoor space, so why not add a few pots of flowers to your porch or patio this summer?", Format: "paragraph"},
		},
	},
}

func PickArticles() T {
	return Articles[rand.Intn(len(Articles))]
}
func PickQuotes() string {
	return quotes[rand.Intn(len(quotes))]
}

func PickImage() string {
	return images[rand.Intn(len(images))]

}
func TestTemplate(c *gin.Context) {

	c.HTML(http.StatusOK, "index.tmpl.html", PayloadArticleContent{
		PublisherName:      "IlliPress",
		PublisherLogo:      "https://s3.us-east-1.amazonaws.com/cdglb-content-server/img/sites/J2CPaXFfi/5gEiWdJKZ_full.jpg",
		ArticleTitle:       PickQuotes(),
		ArticleDescription: PickQuotes(),
		Author:             "Magic CoNdigolabs ",
		MainImage:          PickImage(),
		BreadCrumb: []Action{{
			Name: "Apparel & Accessories",
			Link: "#",
		}, {
			Name: "Clothing",
			Link: "#",
		},
			{
				Name: "Shirts & Tops",
				Link: "#",
			}},
		Quote: Item{
			Title:    PickQuotes(),
			HeadLine: PickQuotes(),
		},
		Articles: []Item{
			{
				Id:       "placement-1",
				Title:    PickQuotes(),
				Image:    PickImage(),
				HeadLine: PickQuotes(),
				Lines:    PickArticles().Lines,
				Actions:  nil,
				Ads: models.BidResponse{
					Id:        "placement-1",
					RequestID: "46db40be-ae5f-4de0-9d33-16d877be8005",
					Country:   "USA",
					Brand:     "Fashion Nova",
					LabelId:   125,
					DATE:      time.Time{},
					Locale:    "en-US",
					Label:     "Apparel & Accessories > Clothing > Dresses",
					Products: []models.Products{{
						Title:       "Womens Madeline Embellished Maxi Dress in Black Size XS by Fashion Nova\"",
						Description: "Available In Black. Sequin Maxi Dress Deep V Neckline Long Sleeve Cut Out Back Zipper Train Lined Stretch Self: 95% Polyester, 5% Spandex Lining: 100% Polyester Imported | Madeline Embellished Maxi Dress in Black size XS by Fashion Nova",

						Domain: "fashionnova.com",
						Image:  "https://pix.us.criteo.net/img/img?c=3&cq=256&h=800&m=0&partner=23261&q=80&r=0&u=https%3A%2F%2Fcdn.shopify.com%2Fs%2Ffiles%2F1%2F0293%2F9277%2Fproducts%2F08-21-19_MS_HOLIDAY_11822_RG.jpg%3Fv%3D1648771543&ups=1&v=3&w=800&s=yaut253JZRsZoA5aXJ-j6QGN",
						Target: "https://www.fashionnova.com/products/madeline-embellished-maxi-dress-black?variant=12204970836092&utm_source=criteoARO&utm_medium=display&utm_campaign=Web%20Conversion%20-%20Jul%2021,%202020",
					},
						{
							Title:       "Womens Anastasia Embellished Maxi Dress in Emerald Size 2X by Fashion Nova",
							Description: "Available In White And Emerald. Maxi Dress Sleeveless Asymmetrical Neckline High Slit Crystal Embellished Detail Lace Up Back Zipper Fabric Content: 95% Polyester 5% Spandex Lining: 95% Polyester 5% Spandex Imported | Anastasia Embellished Maxi Dress in Emerald size 2X by Fashion Nova",
							Domain:      "fashionnova.com",
							Image:       "https://pix.us.criteo.net/img/img?c=3&cq=256&h=800&m=0&partner=23261&q=80&r=0&u=https%3A%2F%2Fcdn.shopify.com%2Fs%2Ffiles%2F1%2F0293%2F9277%2Fproducts%2F11-04-21Studio6_SN_RL_11-09-52_8_A81949_Emerald_4207_PB.jpg%3Fv%3D1649118386&ups=1&v=3&w=800&s=7uykgXcQ6LXHHGnupN6o3PWg",
							Target:      "https://www.fashionnova.com/products/anastasia-embellished-maxi-dress-emerald?variant=39249783160956&utm_source=criteoARO&utm_medium=display&utm_campaign=Web%20Conversion%20-%20Jul%2021,%202020",
						},
					},
				},
			},
			{
				Id:       "placement-2",
				Title:    PickQuotes(),
				Image:    PickImage(),
				HeadLine: PickQuotes(),
				Lines:    PickArticles().Lines,
				Ads: models.BidResponse{
					RequestID: "46db40be-ae5f-4de0-9d33-16d877be8005",
					Id:        "placement-2",
					Brand:     "Fashion Nova",
					Country:   "USA",
					LabelId:   5250,
					Locale:    "en-US",
					Label:     "Apparel & Accessories > Clothing > One-Pieces > Jumpsuits & Rompers",
					Products: []models.Products{
						{
							Title:       "Womens Jasmine Floral Applique Jumpsuit in White Size XS by Fashion Nova",
							Description: "Available In Black And White. Jumpsuit Sleeveless Bustier Top Ruffle Detail Floral Applique Flare Leg Zipper Self 1 92% Polyester 8% Spandex Self 2 95% Polyester 5% Spandex Contrast 100% Polyester Imported | Jasmine Floral Applique Jumpsuit in White size XS by Fashion Nova",

							Domain: "fashionnova.com",
							Image:  "https://pix.us.criteo.net/img/img?c=3&cq=256&h=800&m=0&partner=23261&q=80&r=0&u=https%3A%2F%2Fcdn.shopify.com%2Fs%2Ffiles%2F1%2F0293%2F9277%2Fproducts%2F12-15-21Studio3_ME_KG_14-26-11_44_B02203_White_19531_PB.jpg%3Fv%3D1648774653&ups=1&v=3&w=800&s=Eq1f3QWMHQLPTBvXeOSe7b_C",
							Target: "https://www.fashionnova.com/products/jasmine-floral-applique-jumpsuit-white?variant=39249782505596&utm_source=criteoARO&utm_medium=display&utm_campaign=Web%20Conversion%20-%20Jul%2021,%202020",
						},
					},
				},
			},
			{
				Id:       "placement-3",
				Title:    PickQuotes(),
				Image:    PickImage(),
				HeadLine: PickQuotes(),
				Lines:    PickArticles().Lines,
			},
		},
	})
}
