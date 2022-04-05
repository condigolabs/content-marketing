package intent

import (
	"cloud.google.com/go/bigquery"
	"context"
	"github.com/condigolabs/content-marketing/models"
	"github.com/sirupsen/logrus"
	"time"
)

func (dw *ConcreteIntent) CreateSchema(f models.BqTable) error {
	ctx := context.Background()
	s, err := bigquery.InferSchema(f)
	if err == nil {
		tableName := f.GetTable()
		metaData := &bigquery.TableMetadata{
			Schema:         s,
			ExpirationTime: time.Now().AddDate(5, 0, 0), // Table will be automatically deleted in 5 year.
		}
		if f.IsPartition() {
			metaData.TimePartitioning = &bigquery.TimePartitioning{
				Field:      "Date",
				Expiration: 30 * 24 * time.Hour,
			}
			metaData.RequirePartitionFilter = true
		}
		tableRef := dw.client.Dataset(f.GetDataSet()).Table(tableName)
		if err := tableRef.Create(ctx, metaData); err != nil {
			return err
		}
	} else {
		logrus.WithError(err).Errorf("Creating Schema")
	}
	return nil
}
func (dw *ConcreteIntent) FlushEntitySync(model models.BqTable) error {
	return dw.innerFlushEntity(model)
}

func (dw *ConcreteIntent) innerFlushEntity(model models.BqTable) error {
	start := time.Now()
	ctx := context.Background()
	u := dw.client.Dataset(model.GetDataSet()).Table(model.GetTable()).Inserter()
	items := make([]*bigquery.StructSaver, 0)
	if model.Parse() {
		model.SetRunTime(time.Now().Unix())
		sss := bigquery.StructSaver{
			Schema:   nil,
			InsertID: model.GetId(),
			Struct:   model,
		}
		items = append(items, &sss)
		//Flush All
		if err := u.Put(ctx, items); err != nil {
			logrus.WithError(err).Errorf("inserting innerFlushEntity")
			return err
		}
		elapsed := time.Since(start)
		logrus.Infof("Flush %s %d", model.GetTable(), elapsed.Milliseconds())
	}
	return nil
}

func (dw *ConcreteIntent) BackgroundWorker(queue <-chan models.BqTable) {
	logrus.Infof("BackgroundWorker OK->")
	for p := range queue {
		_ = dw.innerFlushEntity(p)
	}
}
