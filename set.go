package gomongo

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Set = Set struct
type Set struct {
	tableName      string
	result         interface{}
	gom            *Gomongo
	filter         interface{}
	pipe           []bson.M
	sortField      *string
	sortBy         *int
	skip           *int
	limit          *int
	command        *Command
	contextTimeout time.Duration
}

// newSet = init new set
func newSet(gom *Gomongo, params *SetParams) *Set {
	s := new(Set)
	if params == nil {
		s.filter = bson.M{}
		s.pipe = nil
		s.skip = nil
		s.limit = nil
		s.result = nil
		s.tableName = ""
		s.sortField = nil
		s.sortBy = nil
		s.contextTimeout = 30
	} else {
		s.filter = bson.M{}
		if params.Filter != nil {
			s.Filter(params.Filter)
		}

		if params.Pipe != nil {
			s.Pipe(params.Pipe)
		}

		if params.Skip != 0 {
			s.Skip(params.Skip)
		}

		if params.Limit != 0 {
			s.Limit(params.Limit)
		}

		if params.Result != nil {
			s.Result(params.Result)
		}

		if params.TableName != "" {
			s.Table(params.TableName)
		}

		if params.SortField != "" {
			s.Sort(params.SortField, params.SortBy)
		}

		if params.Timeout == 0 {
			s.Timeout(30)
		} else {
			s.Timeout(params.Timeout)
		}
	}

	s.gom = gom
	s.command = newCommand(s)

	return s
}

func (s *Set) reset() {
	s.filter = bson.M{}
	s.limit = nil
	s.pipe = nil
	s.result = nil
	s.skip = nil
	s.sortBy = nil
	s.sortField = nil
	s.tableName = ""
}

// Table = set table/collection name
func (s *Set) Table(tableName string) *Set {
	s.tableName = tableName

	return s
}

// Cmd = choose Command
func (s *Set) Cmd() *Command {
	return s.command
}

// Result = set target of result
func (s *Set) Result(result interface{}) *Set {
	s.result = result

	return s
}

// Skip = set skip data
func (s *Set) Skip(skip int) *Set {
	s.skip = &skip

	return s
}

// Limit = set limit data
func (s *Set) Limit(limit int) *Set {
	s.limit = &limit

	return s
}

// Sort = set sort data
func (s *Set) Sort(field, sortBy string) *Set {
	s.sortField = &field
	sort := -1

	if strings.ToLower(sortBy) == "asc" {
		sort = 1
	}

	s.sortBy = &sort

	return s
}

// Filter = set filter data
func (s *Set) Filter(filter *Filter) *Set {

	if filter != nil {
		main := BuildFilter(filter)
		s.filter = main
	} else {
		s.filter = bson.M{}
	}

	return s
}

// Pipe = set pipe, if this is set => Filter will be ignored
func (s *Set) Pipe(pipe []bson.M) *Set {
	s.pipe = pipe

	return s
}

func (s *Set) buildPipe() []bson.M {
	pipe := []bson.M{}

	if s.pipe != nil {
		pipe = s.pipe
	} else {
		if s.filter != nil {
			pipe = append(pipe, bson.M{
				"$match": s.filter.(bson.M),
			})
		}
	}

	if s.skip != nil {
		pipe = append(pipe, bson.M{
			"$skip": s.skip,
		})
	}

	if s.limit != nil {
		pipe = append(pipe, bson.M{
			"$limit": s.limit,
		})
	}

	if s.sortField != nil {
		pipe = append(pipe, bson.M{
			"$sort": bson.M{
				*s.sortField: s.sortBy,
			},
		})
	}

	return pipe
}

func getValidID(key string) string {
	if key == "ID" || key == "_id" || key == "id" {
		return "_id"
	}

	return key
}

func validateJSONRaw(k string, v json.RawMessage, m bson.M) {
	s := string(v)

	i, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		m[getValidID(k)] = i
		return
	}
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		m[getValidID(k)] = f
		return
	}
	var t time.Time
	err = json.Unmarshal(v, &t)
	if err == nil {
		m[getValidID(k)] = t
		return
	}
	// 26 => includes double quotes
	if len(s) == 26 {
		var oid primitive.ObjectID
		err = json.Unmarshal(v, &oid)
		if err == nil {
			m[getValidID(k)] = oid
			return
		}
	}
	var objMap map[string]json.RawMessage
	err = json.Unmarshal(v, &objMap)
	if err == nil {
		objMapToBsonM := bson.M{}
		for ko, vo := range objMap {
			validateJSONRaw(ko, vo, objMapToBsonM)
		}

		m[getValidID(k)] = objMapToBsonM
		return
	}
	var slice []json.RawMessage
	err = json.Unmarshal(v, &slice)
	if err == nil {
		tempBsonM := bson.M{}
		validSlice := []interface{}{}
		for _, elSlice := range slice {
			validateJSONRaw(RandomString(32), elSlice, tempBsonM)
		}
		for _, vo := range tempBsonM {
			validSlice = append(validSlice, vo)
		}

		m[getValidID(k)] = validSlice
		return
	}
	var itf interface{}
	err = json.Unmarshal(v, &itf)
	if err == nil {
		m[getValidID(k)] = itf
		return
	}
	m[getValidID(k)] = v
}

// buildData = buildData from struct/map to bson M
func (s *Set) buildData(data interface{}, includeID bool) (interface{}, error) {
	var result interface{}
	dataM := bson.M{}

	rv := reflect.ValueOf(data)

	if rv.Kind() != reflect.Ptr {
		return nil, errors.New("data argument must be pointer")
	}

	switch rv.Elem().Kind() {
	case reflect.Struct:
		s, _ := json.Marshal(rv.Interface())

		var mRaw map[string]json.RawMessage

		json.Unmarshal(s, &mRaw)

		for k, v := range mRaw {
			if includeID {
				validateJSONRaw(k, v, dataM)
			} else {
				if k != "_id" {
					validateJSONRaw(k, v, dataM)
				}
			}
		}
		result = dataM

	case reflect.Map:
		v := reflect.ValueOf(rv.Elem().Interface())

		for _, key := range v.MapKeys() {
			value := v.MapIndex(key)
			if includeID {
				dataM[getValidID(key.String())] = value.Interface()
			} else {
				if key.String() != "_id" {
					dataM[getValidID(key.String())] = value.Interface()
				}
			}
		}

		result = dataM

	case reflect.Slice:
		v := reflect.ValueOf(rv.Elem().Interface())

		datas := make([]interface{}, 0)
		for i := 0; i < v.Len(); i++ {
			value := v.Index(i).Interface()
			datas = append(datas, value)
		}

		result = datas

	default:
		return nil, errors.New("data argument must be a struct or map")
	}

	if result == nil {
		return nil, errors.New("data argument can't be empty")
	}

	return result, nil
}

// Timeout = Timeout for command
func (s *Set) Timeout(seconds time.Duration) *Set {
	if &seconds == nil {
		seconds = 30
	}

	s.contextTimeout = seconds

	return s
}

// GetContext = GetContext for command
func (s *Set) GetContext() (context.Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), s.contextTimeout*time.Second)

	return ctx, cancelFunc
}

func RandomString(length int) string {
	return GenerateRandomString("", length)
}

func GenerateRandomString(baseChars string, n int) string {
	if baseChars == "" {
		baseChars = "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnpqrstuvwxyz@#!"
	}
	baseCharsLen := len(baseChars)

	rnd := ""
	for i := 0; i < n; i++ {
		x := RandInt(baseCharsLen)
		rnd += string(baseChars[x])
	}
	return rnd
}
