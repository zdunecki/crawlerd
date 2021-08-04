package mgostorage

import (
	"context"
	"time"

	"crawlerd/pkg/meta/v1"
	"crawlerd/pkg/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type historyrepo struct {
	coll *mongo.Collection
}

func NewHistoryRepository(coll *mongo.Collection) storage.HistoryRepository {
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

func (h *historyrepo) FindByID(ctx context.Context, id int) ([]v1.History, error) {
	cursor, err := h.coll.Find(ctx, bson.M{
		"id": id,
	})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var histories []v1.History

	for cursor.Next(ctx) {
		var historyDocument v1.History

		if err := cursor.Decode(&historyDocument); err != nil {
			return nil, err
		}

		histories = append(histories, v1.History{
			Response:        historyDocument.Response,
			CreatedAt:       historyDocument.CreatedAt,
			DurationSeconds: historyDocument.DurationSeconds,
		})
	}

	return histories, nil
}
