package mgostorage

import (
	"context"
	"time"

	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/objects"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type historyrepo struct {
	coll *mongo.Collection
}

func NewHistoryRepository(coll *mongo.Collection) storage.RepositoryHistory {
	return &historyrepo{
		coll: coll,
	}
}

func (h *historyrepo) InsertOne(ctx context.Context, id int, response []byte, duration time.Duration, createdAt time.Time) (bool, int, error) {
	result, err := h.coll.InsertOne(ctx, bson.M{
		"id":         id,
		"response":   string(response),
		"duration":   duration.Seconds(),
		"created_at": createdAt.Unix(),
	})

	if err != nil {
		return false, 0, err
	}

	if result.InsertedID == nil {
		return false, 0, nil
	}

	return true, id, nil
}

func (h *historyrepo) FindByID(ctx context.Context, id int) ([]objects.History, error) {
	cursor, err := h.coll.Find(ctx, bson.M{
		"id": id,
	})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var histories []objects.History

	for cursor.Next(ctx) {
		var historyDocument objects.History

		if err := cursor.Decode(&historyDocument); err != nil {
			return nil, err
		}

		histories = append(histories, objects.History{
			Response:        historyDocument.Response,
			CreatedAt:       historyDocument.CreatedAt,
			DurationSeconds: historyDocument.DurationSeconds,
		})
	}

	return histories, nil
}
