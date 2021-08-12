package mgostore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type client struct {
	dbName string
	client *mongo.Client
}

func NewClient(opts ...*options.ClientOptions) (*client, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	var passOpts []*options.ClientOptions

	passOpts = append(passOpts, options.Client().ApplyURI(DefaultMongoAddr))
	for _, o := range opts {
		passOpts = append(passOpts, o)
	}

	c, err := mongo.Connect(ctx, passOpts...)

	if err != nil {
		return nil, err
	}

	return &client{
		client: c,
		dbName: DefaultDatabaseName,
	}, nil
}

func (c *client) SetDatabaseName(dbName string) {
	c.dbName = dbName
}

func (c *client) DB() *mongo.Database {
	return c.client.Database(DefaultDatabaseName)
}
