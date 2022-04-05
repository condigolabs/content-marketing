package intent

import (
	"github.com/condigolabs/content-marketing/models"
	"io"
)

type Intent interface {
	io.Closer
	/*
	   Flush and Schema
	*/
	CreateSchema(table models.BqTable) error
	FlushEntitySync(model models.BqTable) error

	CreateIntentTable(p Param) (string, error)
	LoadLatestIntent(p Param) (*models.Intent, error)
	LoadLatestProducts(p Param) ([]models.LatestProduct, error)
	LoadLatestDomain(p Param) ([]models.LatestDomain, error)
	GenerateDocument(p DocHeadlines) (models.GenerateData, error)
	LoadGenerated(requestId string) ([]models.GenerateData, error)
	LoadRequest(p Param) ([]models.Request, error)
	LoadRandImage(p Param) ([]models.ImageIntent, error)
}
