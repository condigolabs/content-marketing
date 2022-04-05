package intent

import (
	"bytes"
	"cloud.google.com/go/bigquery"
	"context"
	"fmt"
	"github.com/condigolabs/content-marketing/models"
	"github.com/nlpcloud/nlpcloud-go"
	_ "github.com/nlpcloud/nlpcloud-go"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"net/http"
	_ "net/http"
	"strings"
)

type DocHeadlines struct {
	Labels []string
}

func (dw *ConcreteIntent) createClient() (*nlpcloud.Client, string, string) {
	model := "finetuned-gpt-neox-20b"
	lang := "fr"
	return nlpcloud.NewClient(&http.Client{}, model, "6945baaef867c37430dfbf92045f9bb6485ab73c", true, lang), model, lang

}
func (dw *ConcreteIntent) LoadGenerated(requestId string) ([]models.GenerateData, error) {
	var buf bytes.Buffer
	err := dw.templates.ExecuteTemplate(&buf, "load_generated_request.sql", requestId)

	if err != nil {
		return nil, err
	}
	fmt.Printf("***\n")
	fmt.Printf(buf.String())
	q := dw.client.Query(buf.String())
	q.Priority = bigquery.InteractivePriority
	q.QueryConfig.UseLegacySQL = false
	ctx := context.Background()
	// Start the job.
	job, err := q.Run(ctx)
	if err != nil {
		return nil, err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return nil, err
	}
	if err := status.Err(); err != nil {
		return nil, err
	}
	ret := make([]models.GenerateData, 0)
	it, err := job.Read(ctx)
	for {
		var row models.GenerateData
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}
		ret = append(ret, row)
	}
	return ret, err
}

func (dw *ConcreteIntent) GenerateDocument(p DocHeadlines) (models.GenerateData, error) {
	/*dfbf92045f9bb6485ab73c
	lang: fr
	model: finetuned-gpt-neox-20b
	keywords: Des vacances  au soleil
	*/

	//k1, _ := keywords.Extract(p.Title)

	client, model, lang := dw.createClient()

	if len(p.Labels) > 10 {
		p.Labels = p.Labels[:10]
	}
	var builder strings.Builder
	for _, s := range p.Labels {
		builder.WriteString(s)
		builder.WriteString("\n")
	}
	r, err := client.AdGeneration(nlpcloud.AdGenerationParams{
		Keywords: p.Labels,
	})
	if err != nil {
		logrus.WithError(err).Errorf("Failed to generate documents")
		return models.GenerateData{}, err
	}
	logrus.Infof("Generated  {%s} ", r.GeneratedText)
	return models.GenerateData{
		InputText:   p.Labels,
		Model:       model,
		Language:    lang,
		Method:      "AdGeneration",
		Description: r.GeneratedText,
	}, nil

}
