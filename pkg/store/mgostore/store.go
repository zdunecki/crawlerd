package mgostore

import (
	"context"

	"crawlerd/pkg/runner"
	"crawlerd/pkg/store"
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

	requestQueue store.RequestQueue
	linker       store.Linker
	urlrepo      store.URL
	historyrepo  store.History
	registryrepo store.Registry
	jobrepo      store.Job
	runnerrepo   runner.Runner
}

func NewStore(db *mongo.Database) Storage {
	m := &mgo{
		seq: db.Collection(DefaultCollectionSequenceName),
	}
	m.requestQueue = NewRequestQueueRepository(db.Collection(DefaultCollectionRequestQueue))
	m.linker = NewLinkerRepository(db.Collection(DefaultCollectionLinker))
	m.urlrepo = NewURLRepository(db.Collection(DefaultCollectionURLName), m)
	m.historyrepo = NewHistoryRepository(db.Collection(DefaultCollectionHistoryName))
	m.registryrepo = NewRegistryRepository(db.Collection(DefaultCollectionRegistryName))
	m.jobrepo = NewJobRepository(db.Collection(DefaultCollectionJobName))
	m.runnerrepo = NewRunnerRepository(db.Collection(DefaultCollectionRunnerName))

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
	return m.urlrepo
}

func (m *mgo) History() store.History {
	return m.historyrepo
}

func (m *mgo) Registry() store.Registry {
	return m.registryrepo
}

func (m *mgo) Job() store.Job {
	return m.jobrepo
}

func (m *mgo) Runner() runner.Runner {
	return m.runnerrepo
}
