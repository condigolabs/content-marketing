package intent

import (
	"cloud.google.com/go/bigquery"
	"context"
	"github.com/condigolabs/content-marketing/models"
	"github.com/gobuffalo/packr"
	"github.com/sirupsen/logrus"
	"text/template"
)

type ConcreteIntent struct {
	client    *bigquery.Client
	box       packr.Box
	templates *template.Template
}

func (dw *ConcreteIntent) GetProducts(intentBag string) ([]models.Product, error) {
	return nil, nil
}

func (dw *ConcreteIntent) Close() error {
	if dw.client != nil {
		return dw.client.Close()
	}
	return nil
}

func New() (Intent, error) {
	logrus.Infof("Intents")
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "core-ssp")
	if err != nil {
		logrus.WithError(err).Errorf("Error Opening BigQuery")
		return nil, err
	}
	ret := &ConcreteIntent{
		box:       packr.NewBox("./scripts"),
		templates: template.New("procedures"),
		client:    client,
	}

	//load all Sql Templates
	files := ret.box.List()
	for _, file := range files {
		b, err := ret.box.Find(file)
		if err == nil {
			ret.templates.New(file).Parse(string(b))
		}
	}

	err = ret.CreateSchema(&models.GenerateData{})

	if err != nil {
		logrus.WithError(err).Errorf("error creating table")
	}

	err = ret.CreateSchema(&models.Article{})
	if err != nil {
		logrus.WithError(err).Errorf("error creating table")
	}

	return ret, nil
}
