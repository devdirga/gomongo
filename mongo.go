package gomongo

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ScramSha1   = "SCRAM-SHA-1"
	ScramSha256 = "SCRAM-SHA-256"
	MongoDbCr   = "MONGODB-CR"
	Plain       = "PLAIN"
	GssAPI      = "GSSAPI"
	MongoDbX509 = "MONGODB-X509"
)

type mongoDB struct {
	Config           Config
	Client           *mongo.Client
	Collection       *mongo.Collection
	ConnectionString string
}

type Config struct {
	Username        string
	Password        string
	Host            string
	Port            int
	Database        string
	MaxPol          int
	AuthMechanism   string
	RegistryBuilder bool
}

func (m *mongoDB) SetConfig(c Config) {
	m.Config = c
}

func (m *mongoDB) SetClient() {
	config := m.Config
	connString := fmt.Sprintf("mongodb://%s:%v", config.Host, config.Port)
	if config.Username != "" {
		connString = fmt.Sprintf("mongodb://%s:%s@%s:%v", config.Username, config.Password, config.Host, config.Port)
		if config.AuthMechanism != "" {
			connString = fmt.Sprintf("mongodb+srv://%s:%s@%s/%s?authMechanism=%s", config.Username, config.Password, config.Host, config.Database, config.AuthMechanism)
		}
	}
	clientOptions := options.Client().ApplyURI(connString)
	if config.RegistryBuilder {
		rb := bson.NewRegistryBuilder()
		rb.RegisterTypeMapEntry(bsontype.EmbeddedDocument, reflect.TypeOf(bson.M{}))
		clientOptions.SetRegistry(rb.Build())
	}
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal()
	}
	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal()
	}
	m.ConnectionString = connString
	m.Client = client
}
