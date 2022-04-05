package generator

import (
	"github.com/nlpcloud/nlpcloud-go"
	"net/http"
)

type ConcreteGenerator struct {
	client *nlpcloud.Client
}

func (c *ConcreteGenerator) Close() error {
	if c.client != nil {

	}
	return nil
}

func New() (Generator, error) {

	client := nlpcloud.NewClient(&http.Client{}, "bart-large-cnn", "4eC39HqLyjWDarjtT1zdp7dc", false, "")
	return &ConcreteGenerator{
		client: client,
	}, nil
}
