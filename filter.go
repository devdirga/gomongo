package gomongo

import (
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type FilterOptions string

const (
	// OpAnd is AND
	OpAnd FilterOptions = "$and"
	// OpOr is OR
	OpOr = "$or"
	// OpNot is Not
	OpNot = "$not"
	// OpEq is Equal
	OpEq = "$eq"
	// OpNe is Not Equal
	OpNe = "$ne"
	// OpGte is Greater than or Equal
	OpGte = "$gte"
	// OpGt is Greater than
	OpGt = "$gt"
	// OpLt is Less than
	OpLt = "$lt"
	// OpLte is Less than or equal
	OpLte = "$lte"
	// OpRange is range from until
	OpRange = "$range"
	// OpContains is Contains
	OpContains = "$contains"
	// OpStartWith is Start with
	OpStartWith = "$startwith"
	// OpEndWith is End with
	OpEndWith = "$endwith"
	// OpIn is In
	OpIn = "$in"
	// OpNin is Not in
	OpNin = "$nin"
	// OpSort is Sort
	OpSort = "$sort"
	// OpBetween is Between (Custom)
	OpBetween = "between"
	// OpExists is Exists
	OpExists = "$exists"
	// OpBetweenEq is Between Equal
	OpBetweenEq = "betweenEq"
	// OpRangeEq is Range Equal
	OpRangeEq = "rangeEq"
	// ElemMatch is Elem Match operator
	OpElemMatch = "$elemMatch"
)

type Filter struct {
	Items []*Filter
	Field string
	Op    FilterOptions
	Value interface{}
}

func newFilter(field string, op FilterOptions, v interface{}, items []*Filter) *Filter {
	// in future make sort condition
	f := new(Filter)
	f.Field = field
	f.Op = op
	f.Value = v
	if items != nil {
		f.Items = items
	}
	return f
}

func And(items ...*Filter) *Filter {
	return newFilter("", OpAnd, nil, items)
}

func Or(items ...*Filter) *Filter {
	return newFilter("", OpOr, nil, items)
}

func Sort(field string, sorttype string) *Filter {
	sort := -1
	if strings.ToLower(sorttype) == "asc" {
		sort = 1
	}
	return newFilter(field, OpSort, sort, nil)
}
func Eq(field string, v interface{}) *Filter {
	return newFilter(field, OpEq, v, nil)
}

// Not create new filter with Not operation
func Not(item *Filter) *Filter {
	return newFilter("", OpNot, nil, []*Filter{item})
}

// Ne create new filter with Ne operation
func Ne(field string, v interface{}) *Filter {
	return newFilter(field, OpNe, v, nil)
}

// Gte create new filter with Gte operation
func Gte(field string, v interface{}) *Filter {
	return newFilter(field, OpGte, v, nil)
}

// Gt create new filter with Gt operation
func Gt(field string, v interface{}) *Filter {
	return newFilter(field, OpGt, v, nil)
}

// Lt create new filter with Lt operation
func Lt(field string, v interface{}) *Filter {
	return newFilter(field, OpLt, v, nil)
}

// Lte create new filter with Lte operation
func Lte(field string, v interface{}) *Filter {
	return newFilter(field, OpLte, v, nil)
}

// Range create new filter with Range operation
func Range(field string, from, to interface{}) *Filter {
	f := newFilter(field, OpRange, nil, nil)
	f.Value = []interface{}{from, to}
	return f
}

// Between create new filter with Between operation (Custom)
func Between(field string, gt, lt interface{}) *Filter {
	f := newFilter(field, OpBetween, nil, nil)
	f.Value = []interface{}{gt, lt}
	return f
}

// RangeEq create new filter with Range Equal operation
func RangeEq(field string, from, to interface{}) *Filter {
	f := newFilter(field, OpRangeEq, nil, nil)
	f.Value = []interface{}{from, to}
	return f
}

// BetweenEq create new filter with Between Equal operation (Custom)
func BetweenEq(field string, gte, lte interface{}) *Filter {
	f := newFilter(field, OpBetweenEq, nil, nil)
	f.Value = []interface{}{gte, lte}
	return f
}

// In create new filter with In operation
func In(field string, inValues ...interface{}) *Filter {
	f := new(Filter)
	f.Field = field
	f.Op = OpIn
	f.Value = inValues
	return f
}

// Nin create new filter with Nin operation
func Nin(field string, ninValues ...interface{}) *Filter {
	f := new(Filter)
	f.Field = field
	f.Op = OpNin
	f.Value = ninValues
	return f
}

// Contains create new filter with Contains operation
func Contains(field string, values ...string) *Filter {
	f := new(Filter)
	f.Field = field
	f.Op = OpContains
	f.Value = values
	return f
}

// StartWith create new filter with StartWith operation
func StartWith(field string, value string) *Filter {
	f := new(Filter)
	f.Field = field
	f.Op = OpStartWith
	f.Value = value
	return f
}

// EndWith create new filter with EndWith operation
func EndWith(field string, value string) *Filter {
	f := new(Filter)
	f.Field = field
	f.Op = OpEndWith
	f.Value = value
	return f
}

// Exists match the documents that contain the field
func Exists(field string, value bool) *Filter {
	f := new(Filter)
	f.Field = field
	f.Op = OpExists
	f.Value = value
	return f
}

// Exists match the documents that contain the field
func ElemMatch(field string, filter *Filter) *Filter {
	f := new(Filter)
	f.Field = field
	f.Op = OpElemMatch
	f.Value = filter
	return f
}

func BuildFilter(filter *Filter) bson.M {
	main := bson.M{}
	inside := bson.M{}

	switch filter.Op {
	case OpAnd, OpOr:
		insideArr := []interface{}{}

		for _, fi := range filter.Items {
			fRes := BuildFilter(fi)
			insideArr = append(insideArr, fRes)
		}

		main[string(filter.Op)] = insideArr

	case OpEq, OpNe, OpGt, OpGte, OpLt, OpLte, OpIn, OpNin, OpSort, OpExists:
		inside[string(filter.Op)] = filter.Value
		main[filter.Field] = inside

	// case OpSort:
	// 	inside.Set(string(filter.Op), filter.Value)
	// 	main.Set(filter.Field, inside)

	case OpBetween, OpRange:
		switch filter.Value.([]interface{})[0].(type) {
		case int:
			gt := 0
			lt := 0

			if filter.Value != nil {
				gt = filter.Value.([]interface{})[0].(int)
				lt = filter.Value.([]interface{})[1].(int)
			}

			main[filter.Field] = bson.M{
				"$gt": gt,
				"$lt": lt,
			}
		case time.Time:
			gt := time.Now()
			lt := time.Now()

			if filter.Value != nil {
				gt = filter.Value.([]interface{})[0].(time.Time)
				lt = filter.Value.([]interface{})[1].(time.Time)
			}

			main[filter.Field] = bson.M{
				"$gt": gt,
				"$lt": lt,
			}
		}

	case OpBetweenEq, OpRangeEq:
		switch filter.Value.([]interface{})[0].(type) {
		case int:
			gt := 0
			lt := 0

			if filter.Value != nil {
				gt = filter.Value.([]interface{})[0].(int)
				lt = filter.Value.([]interface{})[1].(int)
			}

			main[filter.Field] = bson.M{
				"$gte": gt,
				"$lte": lt,
			}
		case time.Time:
			gt := time.Now()
			lt := time.Now()

			if filter.Value != nil {
				gt = filter.Value.([]interface{})[0].(time.Time)
				lt = filter.Value.([]interface{})[1].(time.Time)
			}

			main[filter.Field] = bson.M{
				"$gte": gt,
				"$lte": lt,
			}
		}

	case OpStartWith:
		main[filter.Field] = bson.M{
			"$regex":   fmt.Sprintf("^%s.*$", filter.Value),
			"$options": "i",
		}

	case OpEndWith:
		main[filter.Field] = bson.M{
			"$regex":   fmt.Sprintf("^.*%s$", filter.Value),
			"$options": "i",
		}

	case OpContains:
		if len(filter.Value.([]string)) > 1 {
			bfs := []interface{}{}
			for _, ff := range filter.Value.([]string) {
				pfm := bson.M{}
				pfm[filter.Field] = bson.M{
					"$regex":   fmt.Sprintf(".*%s.*", ff),
					"$options": "i",
				}

				bfs = append(bfs, pfm)
			}
			main["$or"] = bfs
		} else {
			main[filter.Field] = bson.M{
				"$regex":   fmt.Sprintf(".*%s.*", filter.Value.([]string)[0]),
				"$options": "i",
			}
		}

	case OpNot:
		// field := filter.Items[0].Field
		// main.Set(field, toolkit.M{}.Set("$not", filter.Items[0].Field))
		// toolkit.Println(toolkit.JsonStringIndent(main, "\n"))

	case OpElemMatch:
		inside[string(filter.Op)] = BuildFilter(filter.Value.(*Filter))
		main[filter.Field] = inside

	}

	return main
}
