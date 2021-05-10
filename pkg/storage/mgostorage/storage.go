package mgostorage

import (
	"context"

	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/objects"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type StorageRepository interface {
	InsertedID(name string) (int, error)
}

type Storage interface {
	storage.Client
	StorageRepository
}

type mgo struct {
	seq *mongo.Collection

	urlrepo     storage.RepositoryURL
	historyrepo storage.RepositoryHistory
}

func NewStorage(db *mongo.Database) Storage {
	mongodb := &mgo{
		seq: db.Collection(DefaultCollectionSequenceName),
	}
	mongodb.urlrepo = NewURLRepository(db.Collection(DefaultCollectionURLName), mongodb)
	mongodb.historyrepo = NewHistoryRepository(db.Collection(DefaultCollectionHistoryName))

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

func (m *mgo) URL() storage.RepositoryURL {
	return m.urlrepo
}

func (m *mgo) History() storage.RepositoryHistory {
	return m.historyrepo
}
