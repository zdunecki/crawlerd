package mgostore

import (
	"context"

	metav1 "crawlerd/pkg/meta/v1"
	"crawlerd/pkg/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: logger
type job struct {
	coll         *mongo.Collection
	jobFunctions *jobFunctions
}

func NewJobRepository(coll *mongo.Collection) store.Job {
	j := &job{
		coll: coll,
	}

	j.jobFunctions = NewJobFunctions(j)

	return j
}

func (j *job) FindAll(ctx context.Context) ([]metav1.Job, error) {
	cursor, err := j.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var jobs []metav1.Job

	for cursor.Next(ctx) {
		var job metav1.Job

		if err := cursor.Decode(&job); err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (j *job) FindOneByID(ctx context.Context, id string) (*metav1.Job, error) {
	objID, _ := primitive.ObjectIDFromHex(id)

	cursor := j.coll.FindOne(ctx, bson.M{
		"_id": objID,
	})

	var job metav1.Job

	if err := cursor.Decode(&job); err != nil {
		return nil, err
	}

	return &job, nil
}

func (j *job) InsertOne(ctx context.Context, job *metav1.JobCreate) (string, error) {
	resp, err := j.coll.InsertOne(ctx, job)
	if err != nil {
		return "", err
	}

	if oid, ok := resp.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}

	return "", nil
}

func (j *job) PatchOneByID(ctx context.Context, id string, job *metav1.JobPatch) error {
	objID, _ := primitive.ObjectIDFromHex(id)

	_, err := j.coll.UpdateByID(ctx, objID, bson.M{"$set": job})

	if err != nil {
		return err
	}

	return err
}

func (j *job) Functions() store.Functions {
	return j.jobFunctions
}
