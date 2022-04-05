package controller

import (
	"fmt"
	"github.com/condigolabs/content-marketing/models"
	"github.com/condigolabs/content-marketing/services/intent"
	"github.com/condigolabs/content-marketing/startup"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func InitRouter(engine *gin.Engine) {
	intents := engine.Group("/intents")

	intents.GET("bids", GetLatestProducts)
	intents.GET("intents", GetLatestIntent)
	intents.GET("request/:id/:ext", GetRequest)
	intents.GET("cat/:tag", GetRequest)
}

type Dimensions struct {
	Type   string   `json:"type"`
	Label  string   `json:"label"`
	Values []string `json:"values"`
}
type Tag struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}
type LatestProducts struct {
	Dimensions []Dimensions           `json:"dimensions"`
	Bids       []models.LatestProduct `json:"bids"`
	Tags       []Tag                  `json:"tags"`
}

func GetLatestProducts(c *gin.Context) {
	service := startup.GetIntent()
	param := intent.Param{
		Country:   "FRA",
		LastHours: 120,
		Locale:    "fr-FR",
	}
	ret, err := service.LoadLatestProducts(param)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}

	for i := 0; i < len(ret); i++ {
		ret[i].GenerateLink = fmt.Sprintf("http://api.condigolabs.com:8080/intents/request/%s", ret[i].RequestId)
	}
	tags, err := service.LoadLatestIntent(param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	domain, err := service.LoadLatestDomain(param)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	vTags := make([]Tag, len(tags.Labels))
	for i := 0; i < len(vTags); i++ {
		vTags[i] = Tag{
			Value: tags.Labels[i].Cat,
			Count: tags.Labels[i].Count,
		}
	}
	d := make([]string, len(domain))
	for i := 0; i < len(d); i++ {
		d[i] = domain[i].Domain
	}

	c.JSON(http.StatusOK, LatestProducts{
		Tags: vTags,
		Dimensions: []Dimensions{{
			Type:   "domain",
			Label:  "domain",
			Values: d,
		},
			{
				Type:   "country",
				Label:  "country",
				Values: []string{"FRA", "USA"},
			},
		},
		Bids: ret,
	})
}

func GetLatestIntent(c *gin.Context) {
	service := startup.GetIntent()
	ret, err := service.LoadLatestIntent(intent.Param{
		Country:   "FRA",
		LastHours: 120,
		Locale:    "fr-FR",
	})
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
	//tag := c.Param("tag")
	var ret []models.GenerateData

	if len(requestId) > 0 {
		request, err := service.LoadRequest(intent.Param{
			Locale:    "fr-FR",
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
