package reader

import (
	"github.com/condigolabs/content-marketing/models"
	"github.com/go-shiori/go-readability"
	"github.com/sirupsen/logrus"
	"net/url"
	"time"
)

func Do(uri string, labelId int64) (models.Article, error) {
	u, _ := url.Parse(uri)

	article, err := readability.FromURL(u.String(), 30*time.Second)
	if err != nil {
		logrus.WithError(err).Errorf("failed")
		return models.Article{
			Date:    time.Now(),
			Domain:  u.Host,
			Uri:     u.String(),
			Status:  0,
			LabelId: labelId}, err
	}
	return models.Article{
		Date:        time.Now(),
		Domain:      u.Host,
		Uri:         u.String(),
		Status:      1,
		Title:       article.Title,
		Byline:      article.Byline,
		Content:     article.Content,
		TextContent: article.TextContent,
		Length:      article.Length,
		Excerpt:     article.Excerpt,
		SiteName:    article.SiteName,
		Image:       article.Image,
		Favicon:     article.Favicon,
		LabelId:     labelId,
	}, err
}
