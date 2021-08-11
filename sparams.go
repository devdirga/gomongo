package gomongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type SetParams struct {
	TableName string
	Result    interface{}
	Filter    *Filter
	Pipe      []bson.M
	SortField string
	SortBy    string
	Skip      int
	Limit     int
	Timeout   time.Duration
}

// NewSetParams = Init set params
func (g *Gomongo) NewSetParams() *SetParams {
	sp := new(SetParams)
	sp.Timeout = 30

	return sp
}
