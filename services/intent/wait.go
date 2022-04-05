package intent

import (
	"cloud.google.com/go/bigquery"
	"context"
	"github.com/sirupsen/logrus"
)

func RunAndWait(ctx context.Context, q *bigquery.Query) (*bigquery.Job, error) {
	job, err := q.Run(ctx)
	if err != nil {
		logrus.WithError(err).Infof("Error Running")
		return nil, err
	}
	return wait1(ctx, job)
}

func wait1(ctx context.Context, job *bigquery.Job) (*bigquery.Job, error) {

	status, err := job.Wait(ctx)
	if err != nil {
		logrus.WithError(err).Errorf("Error waiting ")
		return job, err
	}

	if err := status.Err(); err != nil {
		logrus.WithError(err).Errorf("status= %d", status.State)
		return job, err
	}
	logrus.Infof("status=%d, done=%t bytes=%d", status.State, status.Done(),
		status.Statistics.Details, status.Statistics.TotalBytesProcessed)
	return job, nil
}
