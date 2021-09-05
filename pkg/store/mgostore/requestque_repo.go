package mgostore

import (
	"context"
	"encoding/json"

	metav1 "crawlerd/pkg/meta/v1"
	"crawlerd/pkg/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: logger
type requestQueue struct {
	coll *mongo.Collection
}

func NewRequestQueueRepository(coll *mongo.Collection) store.RequestQueue {
	rq := &requestQueue{
		coll: coll,
	}

	return rq
}

func (rq *requestQueue) List(ctx context.Context, filter *metav1.RequestQueueListFilter) ([]*metav1.RequestQueue, error) {
	query := bson.M{}

	var queryFilters map[string]interface{}

	filterB, _ := json.Marshal(filter)
	if err := json.Unmarshal(filterB, &queryFilters); err != nil {
		return nil, err
	}

	for key, f := range queryFilters {
		switch filter := f.(type) {
		case *metav1.StringFilter:
			query[key] = filter.Is
		case *metav1.UintFilter:
			query[key] = filter.Is
		}
	}

	// TODO: translate query to mongo
	cursor, err := rq.coll.Find(ctx, query)
	if err != nil {
		return nil, err
	}

	var nodes []*metav1.RequestQueue

	for cursor.Next(ctx) {
		var node metav1.RequestQueue

		if err := cursor.Decode(&node); err != nil {
			return nil, err
		}

		nodes = append(nodes, &node)
	}

	return nodes, err
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

func (rq *requestQueue) UpdateByID(ctx context.Context, id string, patch *metav1.RequestQueuePatch) error {
	objID, _ := primitive.ObjectIDFromHex(id)

	_, err := rq.coll.UpdateByID(ctx, objID, bson.M{"$set": patch})
	return err
}
