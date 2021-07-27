package mgostorage

import (
	"context"

	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/objects"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoStorage interface {
	InsertedID(name string) (int, error)
}

type Storage interface {
	storage.Storage
	MongoStorage
}

type mgo struct {
	seq *mongo.Collection

	urlrepo      storage.URLRepository
	historyrepo  storage.HistoryRepository
	registryrepo storage.RegistryRepository
}

func NewStorage(db *mongo.Database) Storage {
	mongodb := &mgo{
		seq: db.Collection(DefaultCollectionSequenceName),
	}
	mongodb.urlrepo = NewURLRepository(db.Collection(DefaultCollectionURLName), mongodb)
	mongodb.historyrepo = NewHistoryRepository(db.Collection(DefaultCollectionHistoryName))
	mongodb.registryrepo = NewRegistryRepository(db.Collection(DefaultCollectionRegistryName))

	return mongodb
}

func (m *mgo) InsertedID(name string) (int, error) {
	var seq objects.Sequence

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

func (m *mgo) URL() storage.URLRepository {
	return m.urlrepo
}

func (m *mgo) History() storage.HistoryRepository {
	return m.historyrepo
}

func (m *mgo) Registry() storage.RegistryRepository {
	return m.registryrepo
}
