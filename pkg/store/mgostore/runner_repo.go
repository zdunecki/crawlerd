package mgostore

import (
	"context"

	metav1 "github.com/zdunecki/crawlerd/pkg/meta/metav1"
	"github.com/zdunecki/crawlerd/pkg/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: logger
type runnerRepo struct {
	coll *mongo.Collection
}

func NewRunnerRepository(coll *mongo.Collection) store.Runner {
	r := &runnerRepo{
		coll: coll,
	}

	return r
}

func (r *runnerRepo) List(ctx context.Context) ([]*metav1.Runner, error) {
	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var runners []*metav1.Runner

	for cursor.Next(ctx) {
		var runner metav1.Runner

		if err := cursor.Decode(&runner); err != nil {
			return nil, err
		}

		runners = append(runners, &runner)
	}

	return runners, nil
}

func (r *runnerRepo) GetByID(ctx context.Context, id string) (*metav1.Runner, error) {
	objID, _ := primitive.ObjectIDFromHex(id)

	result := r.coll.FindOne(ctx, bson.M{
		"_id": objID,
	})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var runner *metav1.Runner

	if err := result.Decode(&runner); err != nil {
		return nil, err
	}

	return runner, nil
}

func (r *runnerRepo) Create(ctx context.Context, create *metav1.RunnerCreate) (string, error) {
	result, err := r.coll.InsertOne(ctx, create)
	if err != nil {
		return "", err
	}

	id := ""

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		id = oid.Hex()
	}

	return id, nil
}

func (r *runnerRepo) UpdateByID(ctx context.Context, id string, patch *metav1.RunnerPatch) error {
	objID, _ := primitive.ObjectIDFromHex(id)

	_, err := r.coll.UpdateByID(ctx, objID, bson.M{"$set": patch})
	return err
}
