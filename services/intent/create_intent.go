package intent

import (
	"bytes"
	"cloud.google.com/go/bigquery"
	"context"
	"fmt"
	"github.com/condigolabs/content-marketing/models"
	"google.golang.org/api/iterator"
	"time"
)

type Param struct {
	Country   string
	LastHours int
	Locale    string
	Value     string
	RequestId string
	Tag       string
}

func (dw *ConcreteIntent) CreateIntentTable(p Param) (string, error) {

	var buf bytes.Buffer

	err := dw.templates.ExecuteTemplate(
		&buf, "generate_intent.sql", p)

	if err != nil {
		return "", err
	}
	table := "content-top"
	fmt.Printf(buf.String())
	q := dw.client.Query(buf.String())
	q.Priority = bigquery.InteractivePriority
	q.QueryConfig.Dst = dw.client.Dataset("machinelearning").Table("content-top")
	q.QueryConfig.AllowLargeResults = true
	q.QueryConfig.UseLegacySQL = false
	q.QueryConfig.WriteDisposition = bigquery.WriteTruncate

	_, err = RunAndWait(context.Background(), q)
	return table, err
}

func (dw *ConcreteIntent) LoadIntent(p Param) (models.Intent, error) {
	var buf bytes.Buffer
	err := dw.templates.ExecuteTemplate(&buf, "select_publisher_default.sql", p)

	if err != nil {
		return models.Intent{}, err
	}
	fmt.Printf(buf.String())
	q := dw.client.Query(buf.String())
	q.Priority = bigquery.InteractivePriority
	q.QueryConfig.UseLegacySQL = false
	ctx := context.Background()
	// Start the job.
	job, err := q.Run(ctx)
	if err != nil {
		return models.Intent{}, err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return models.Intent{}, err
	}
	if err := status.Err(); err != nil {
		return models.Intent{}, err
	}

	it, err := job.Read(ctx)
	for {
		var row models.LabelIntent
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}
	}

	return models.Intent{}, err
}

func (dw *ConcreteIntent) LoadLatestProducts(p Param) ([]models.LatestProduct, error) {
	var buf bytes.Buffer
	err := dw.templates.ExecuteTemplate(&buf, "load_latest_product.sql", p)

	if err != nil {
		return nil, err
	}
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
	ret := make([]models.LatestProduct, 0)
	it, err := job.Read(ctx)
	for {
		var row models.LatestProduct
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
func (dw *ConcreteIntent) LoadRandImage(p Param) ([]models.ImageIntent, error) {
	var buf bytes.Buffer
	err := dw.templates.ExecuteTemplate(&buf, "load_rand_image.sql", p)

	if err != nil {
		return nil, err
	}
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

	ret := make([]models.ImageIntent, 0)
	it, err := job.Read(ctx)
	for {
		var row models.ImageIntent
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
func (dw *ConcreteIntent) LoadLatestIntent(p Param) (*models.Intent, error) {
	var buf bytes.Buffer
	err := dw.templates.ExecuteTemplate(&buf, "load_latest_intent.sql", p)

	if err != nil {
		return nil, err
	}
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

	ret := models.Intent{
		Date:    time.Now(),
		Country: p.Country,
		Labels:  nil,
	}
	it, err := job.Read(ctx)
	for {
		var row models.LabelIntent
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}
		ret.Append(row)
	}
	return &ret, err
}
func (dw *ConcreteIntent) LoadLatestDomain(p Param) ([]models.LatestDomain, error) {
	var buf bytes.Buffer
	err := dw.templates.ExecuteTemplate(&buf, "get_latest_domain.sql", p)

	if err != nil {
		return nil, err
	}
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

	ret := make([]models.LatestDomain, 0)
	it, err := job.Read(ctx)
	for {
		var row models.LatestDomain
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

func (dw *ConcreteIntent) LoadRequest(p Param) ([]models.Request, error) {
	var buf bytes.Buffer
	err := dw.templates.ExecuteTemplate(&buf, "load_request.sql", p)

	if err != nil {
		return nil, err
	}
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

	ret := make([]models.Request, 0)
	it, err := job.Read(ctx)
	for {
		var row models.Request
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

func (dw *ConcreteIntent) LoadRawData(ctx context.Context, p Param, out chan models.RawData) error {
	var buf bytes.Buffer
	err := dw.templates.ExecuteTemplate(&buf, "load_rawdata.sql", p)

	if err != nil {
		return err
	}
	fmt.Printf(buf.String())
	q := dw.client.Query(buf.String())
	q.Priority = bigquery.InteractivePriority
	q.QueryConfig.UseLegacySQL = false

	// Start the job.
	job, err := q.Run(ctx)
	if err != nil {
		return err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return err
	}
	if err := status.Err(); err != nil {
		return err
	}

	it, err := job.Read(ctx)
	for {
		var row models.RawData
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}
		if err != nil {
			return err
		}
		select {
		case out <- row:
		case <-ctx.Done():
			return nil
		}
	}
	return nil
}
