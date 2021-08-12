package mgostore

import (
	"context"

	"crawlerd/pkg/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoStorage interface {
	InsertedID(name string) (int, error)
}

type Storage interface {
	store.Storage
	MongoStorage
}

type mgo struct {
	seq *mongo.Collection

	requestQueue store.RequestQueueRepository
	linker       store.LinkerRepository
	urlrepo      store.URLRepository
	historyrepo  store.HistoryRepository
	registryrepo store.RegistryRepository
	jobrepo      store.JobRepository
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

func (m *mgo) RequestQueue() store.RequestQueueRepository {
	return m.requestQueue
}

func (m *mgo) Linker() store.LinkerRepository {
	return m.linker
}

func (m *mgo) URL() store.URLRepository {
	return m.urlrepo
}

func (m *mgo) History() store.HistoryRepository {
	return m.historyrepo
}

func (m *mgo) Registry() store.RegistryRepository {
	return m.registryrepo
}

func (m *mgo) Job() store.JobRepository {
	return m.jobrepo
}
