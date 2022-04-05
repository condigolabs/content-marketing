package models

type BqTable interface {
	GetTable() string
	GetDataSet() string
	Parse() bool
	GetId() string
	SetRunTime(runTime int64)
	IsPartition() bool
}
