package gomongo

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Test Commit 20230926-0841

type Gomongo struct {
	mongo mongoDB
}

func NewGomongo() *Gomongo {
	return new(Gomongo)
}

func (g *Gomongo) Initx(c Config) {
	g.mongo.SetConfig(c)
	g.mongo.SetClient()
}

func (g *Gomongo) Set(SetParams *SetParams) *Set {
	s := newSet(g, SetParams)

	return s
}

// CheckClient = Check connection successfull or not
func (g *Gomongo) CheckClient() error {
	err := g.mongo.Client.Ping(context.Background(), readpref.Primary())

	if err != nil {
		return errors.New(fmt.Sprintf("Couldn't connect to database : %s", err.Error()))
	}

	fmt.Println(fmt.Sprintf("Connected to database: %s", g.mongo.ConnectionString))

	return nil
}

// GetClient = Get active client
func (g *Gomongo) GetClient() *mongo.Client {
	return g.mongo.Client
}

// GetDatabase = Get database name
func (g *Gomongo) GetDatabase() string {
	return g.mongo.Config.Database
}
