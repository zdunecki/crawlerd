package mgostore

import (
	"context"

	metav1 "crawlerd/pkg/meta/v1"
	"crawlerd/pkg/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: logger
type requestQueue struct {
	coll *mongo.Collection
}

func NewRequestQueueRepository(coll *mongo.Collection) store.RequestQueueRepository {
	rq := &requestQueue{
		coll: coll,
	}

	return rq
}

func (rq *requestQueue) InsertMany(ctx context.Context, queues []*metav1.RequestQueueCreate) ([]string, error) {
	insertedIDs := make([]string, 0)

	documents := make([]interface{}, len(queues))

	for i, q := range queues {
		documents[i] = q
	}

	manyResult, err := rq.coll.InsertMany(ctx, documents)

	if err != nil {
		return nil, err
	}

	for _, result := range manyResult.InsertedIDs {
		if oid, ok := result.(primitive.ObjectID); ok {
			insertedIDs = append(insertedIDs, oid.Hex())
		}
	}

	return insertedIDs, nil
}
