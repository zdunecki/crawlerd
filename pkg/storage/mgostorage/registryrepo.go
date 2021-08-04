package mgostorage

import (
	"context"
	"encoding/json"
	"time"

	"crawlerd/pkg/meta/v1"
	"crawlerd/pkg/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const minuteTTL = 60
const defaultTTLBuffer = 1 * minuteTTL

type mgoURL struct {
	v1.CrawlURL
	CreatedAt time.Time `json:"created_at"`
}

// TODO: logger
type registry struct {
	coll *mongo.Collection
}

func NewRegistryRepository(coll *mongo.Collection) storage.RegistryRepository {
	return &registry{
		coll: coll,
	}
}

func (r *registry) GetURLByID(ctx context.Context, id int) (*v1.CrawlURL, error) {
	resp := r.coll.FindOne(ctx, bson.M{
		"id": id,
	})

	if resp.Err() != nil {
		return nil, resp.Err()
	}

	var data *mgoURL

	if err := resp.Decode(&data); err != nil {
		return nil, err
	}

	return &data.CrawlURL, nil
}

// TODO: ttl should be from paramter not directly in storage
// TODO: lease ttl bump
func (r *registry) PutURL(ctx context.Context, url v1.CrawlURL) error {
	{
		ttlIndex := mongo.IndexModel{Keys: bson.M{"created_at": 1}, Options: options.Index().SetExpireAfterSeconds(defaultTTLBuffer)}
		_, err := r.coll.Indexes().CreateOne(ctx, ttlIndex)
		if err != nil {
			return err
		}
	}

	{

		data := &mgoURL{
			url,
			time.Now(),
		}

		crawlUrlB, err := json.Marshal(data)
		if err != nil {
			return err
		}

		if _, err := r.coll.InsertOne(ctx, crawlUrlB); err != nil {
			return err
		}
	}

	return nil
}

func (r *registry) DeleteURL(ctx context.Context, url v1.CrawlURL) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{
		"id": url.Id,
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *registry) DeleteURLByID(ctx context.Context, id int) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{
		"id": id,
	})

	if err != nil {
		return err
	}

	return nil
}
