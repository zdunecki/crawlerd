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

func NewLinkerRepository(coll *mongo.Collection) store.LinkerRepository {
	l := &linker{
		coll: coll,
	}

	return l
}

func (l *linker) InsertManyIfNotExists(ctx context.Context, queues []*metav1.RequestQueueCreate) ([]string, error) {
	insertedIDs := make([]string, 0)

	for _, q := range queues {
		find, err := l.coll.Find(ctx, bson.M{
			"url": q.URL,
		}, options.Find().SetLimit(1))

		if err != nil {
			// TODO: log error
			continue
		}

		if find.Next(ctx) {
			continue
		}

		resp, _ := l.coll.InsertOne(ctx, q)

		if oid, ok := resp.InsertedID.(primitive.ObjectID); ok {
			insertedIDs = append(insertedIDs, oid.Hex())
		}
	}

	return insertedIDs, nil
}
