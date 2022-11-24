package mgostore

import (
	"context"

	"github.com/zdunecki/crawlerd/pkg/meta/metav1"
	"github.com/zdunecki/crawlerd/pkg/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	ScrollDefault = 50
)

type seedrepo struct {
	client Storage
	coll   *mongo.Collection

	defaultScroll int
}

func NewSeedRepository(coll *mongo.Collection, client Storage) store.Seed {
	return &seedrepo{
		client:        client,
		coll:          coll,
		defaultScroll: ScrollDefault,
	}
}

func (s *seedrepo) List(ctx context.Context) ([]*metav1.Seed, error) {
	cursor, err := s.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var seed []*metav1.Seed

	for cursor.Next(ctx) {

		var responseURL *metav1.Seed

		if err := cursor.Decode(&responseURL); err != nil {
			return nil, err
		}

		seed = append(seed, responseURL)
	}

	return seed, nil
}

func (u *seedrepo) Append(ctx context.Context, seed []*metav1.Seed) error {
	return nil
	//seq, err := u.client.InsertedID(DefaultCollectionURLName)
	//if err != nil {
	//	return false, 0, err
	//}
	//
	//result, err := u.coll.InsertOne(
	//	ctx,
	//	bson.M{
	//		"id":       seq,
	//		"url":      url,
	//		"interval": interval,
	//	},
	//)
	//
	//if err != nil {
	//	return false, 0, err
	//}
	//
	//if result.InsertedID == nil {
	//	return false, 0, nil
	//}
	//
	//return true, seq, nil
}

func (u *seedrepo) DeleteMany(context.Context, ...string) ([]error, error) {
	return nil, nil
	//result, err := u.coll.DeleteOne(
	//	ctx,
	//	bson.M{
	//		"id": id,
	//	},
	//)
	//if err != nil {
	//	return false, err
	//}
	//
	//if result.DeletedCount == 0 {
	//	return false, nil
	//}
	//
	//return true, nil
}
