package controller

import (
	"context"
	"fmt"
	"github.com/condigolabs/content-marketing/models"
	"github.com/condigolabs/content-marketing/services/intent"
	"github.com/condigolabs/content-marketing/startup"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"net/http"
)

func LoadCategories(c *gin.Context) {

	service := startup.GetIntent()
	out := make(chan models.Categories)

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		defer close(out)
		return service.LoadCategories(ctx,
			intent.Param{
				Country:   "USA",
				LastHours: 120,
				Locale:    "en-US",
			}, out)

	})
	for i := 0; i < 1; i++ {
		for r := range out {
			i, err := GenerateBlogPost(r)
			if err == nil {
				err := service.FlushEntitySync(&i)
				if err != nil {
					logrus.WithError(err).Errorf("error flushing")
				}
			}
		}
		if err := g.Wait(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

/*logrus.Infof("%s", r)
ret, err := unsplash.Search(r.FriendlyName)
if err == nil {
	for _, i := range ret.Results {
		i.LabelId = r.ID
		err := service.FlushEntitySync(&i)
		if err != nil {
			logrus.WithError(err).Errorf("error flushing")
		}
	}
}*/
func GenerateBlogPost(r models.Categories) (models.GeneratedArticle, error) {
	service := startup.GetIntent()

	return service.GenerateArticles(intent.ParamGeneration{
		Model:              "finetuned-gpt-neox-20b",
		Lang:               "",
		Label:              r.Label,
		LabelId:            r.ID,
		InputText:          fmt.Sprintf("generate a blog post about %s", r.FriendlyName),
		MinLength:          100,
		MaxLength:          1000,
		LengthNoInput:      true,
		EndSequence:        "###",
		RemoveInput:        true,
		DoSample:           true,
		NumBeams:           1,
		EarlyStopping:      false,
		NoRepeatNgramSize:  0,
		NumReturnSequences: 1,
		TopK:               50,
		TopP:               0.9,
		Temperature:        0.9,
		RepetitionPenalty:  1.0,
		LengthPenalty:      1.0,
		RemoveEndSequence:  true,
	})
}
