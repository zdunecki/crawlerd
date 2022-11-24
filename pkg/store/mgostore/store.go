package mgostore

import (
	"context"

	"github.com/zdunecki/crawlerd/pkg/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoStorage interface {
	InsertedID(name string) (int, error)
}

type Storage interface {
	store.Repository
	MongoStorage
}

type mgo struct {
	seq *mongo.Collection

	requestQueue       store.RequestQueue
	linker             store.Linker
	urlRepo            store.URL
	historyRepo        store.History
	registryRepo       store.Registry
	jobRepo            store.Job
	runnerRepo         store.Runner
	runnerFunctionRepo store.RunnerFunctions
}

func NewStore(db *mongo.Database) Storage {
	m := &mgo{
		seq: db.Collection(DefaultCollectionSequenceName),
	}
	m.requestQueue = NewRequestQueueRepository(db.Collection(DefaultCollectionRequestQueue))
	m.linker = NewLinkerRepository(db.Collection(DefaultCollectionLinker))
	m.urlRepo = NewURLRepository(db.Collection(DefaultCollectionURLName), m)
	m.historyRepo = NewHistoryRepository(db.Collection(DefaultCollectionHistoryName))
	m.registryRepo = NewRegistryRepository(db.Collection(DefaultCollectionRegistryName))
	m.jobRepo = NewJobRepository(db.Collection(DefaultCollectionJobName))
	m.runnerRepo = NewRunnerRepository(db.Collection(DefaultCollectionRunnerName))
	m.runnerFunctionRepo = NewJobFunctions(m.jobRepo)

	return m
}

func (m *mgo) InsertedID(name string) (int, error) {
	var seq Sequence

	updateSeq := m.seq.FindOneAndUpdate(
		context.Background(),
		bson.M{
			"_id": name,
		},
		bson.M{
			"$inc": bson.M{
				"id": 1,
			},
		},
	)

	err := updateSeq.Err()

	if err == mongo.ErrNoDocuments {
		if _, err := m.seq.InsertOne(context.Background(), bson.M{"_id": name, "id": 0}); err != nil {
			return 0, err
		}

		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	if err := updateSeq.Decode(&seq); err != nil {
		return 0, err
	}

	return seq.ID + 1, nil
}

func (m *mgo) RequestQueue() store.RequestQueue {
	return m.requestQueue
}

func (m *mgo) Linker() store.Linker {
	return m.linker
}

func (m *mgo) URL() store.URL {
	return m.urlRepo
}

func (m *mgo) History() store.History {
	return m.historyRepo
}

func (m *mgo) Registry() store.Registry {
	return m.registryRepo
}

func (m *mgo) Job() store.Job {
	return m.jobRepo
}

func (m *mgo) Runner() store.Runner {
	return m.runnerRepo
}

func (m *mgo) RunnerFunctions() store.RunnerFunctions {
	return m.runnerFunctionRepo
}

func (m *mgo) Seed() store.Seed {
	return nil // TODO:
}
