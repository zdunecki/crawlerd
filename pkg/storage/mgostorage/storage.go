package mgostorage

import (
	"context"

	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/objects"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ClientRepository interface {
	InsertedID(name string) (int, error)
}

type Client interface {
	storage.Client
	ClientRepository
}

type client struct {
	seq *mongo.Collection

	urlrepo     storage.RepositoryURL
	historyrepo storage.RepositoryHistory
}

func NewClient(db *mongo.Database) Client {
	mongodb := &client{
		seq: db.Collection("seq"),
	}
	mongodb.urlrepo = NewURLRepository(db.Collection("urls"), mongodb)
	mongodb.historyrepo = NewHistoryRepository(db.Collection("histories"))

	return mongodb
}

func (m *client) InsertedID(name string) (int, error) {
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

func (m *client) URL() storage.RepositoryURL {
	return m.urlrepo
}

func (m *client) History() storage.RepositoryHistory {
	return m.historyrepo
}
