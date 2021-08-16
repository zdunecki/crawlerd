package mgostore

import (
	"context"

	metav1 "crawlerd/pkg/meta/v1"
	"crawlerd/pkg/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TODO: logger
type linker struct {
	coll *mongo.Collection
}

func NewLinkerRepository(coll *mongo.Collection) store.Linker {
	l := &linker{
		coll: coll,
	}

	return l
}

// TODO: separate structure for queues parameter
func (l *linker) InsertManyIfNotExists(ctx context.Context, queues []*metav1.LinkNodeCreate) ([]string, error) {
	insertedIDs := make([]string, 0)

	for _, q := range queues {
		node := &metav1.LinkNode{
			URL: q.URL,
		}
		find, err := l.coll.Find(ctx, node, options.Find().SetLimit(1))

		if err != nil {
			// TODO: log error
			continue
		}

		if find.Next(ctx) {
			continue
		}

		resp, _ := l.coll.InsertOne(ctx, node)

		if oid, ok := resp.InsertedID.(primitive.ObjectID); ok {
			insertedIDs = append(insertedIDs, oid.Hex())
		}
	}

	return insertedIDs, nil
}

func (l *linker) FindAll(ctx context.Context) ([]*metav1.LinkNode, error) {
	cursor, err := l.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var nodes []*metav1.LinkNode

	for cursor.Next(ctx) {
		var node metav1.LinkNode

		if err := cursor.Decode(&node); err != nil {
			return nil, err
		}

		nodes = append(nodes, &node)
	}

	return nodes, nil
}
