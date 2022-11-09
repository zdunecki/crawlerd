package mgostore

import (
	"context"

	"crawlerd/pkg/meta/metav1"
	"crawlerd/pkg/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
// ScrollDefault = 50
)

type urlrepo struct {
	client Storage
	coll   *mongo.Collection

	defaultScroll int
}

func NewURLRepository(coll *mongo.Collection, client Storage) store.URL {
	return &urlrepo{
		client:        client,
		coll:          coll,
		defaultScroll: ScrollDefault,
	}
}

func (u *urlrepo) FindAll(ctx context.Context) ([]metav1.URL, error) {
	cursor, err := u.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var urls []metav1.URL

	for cursor.Next(ctx) {

		var responseURL metav1.URL

		if err := cursor.Decode(&responseURL); err != nil {
			return nil, err
		}

		urls = append(urls, responseURL)
	}

	return urls, nil
}

func (u *urlrepo) InsertOne(ctx context.Context, url string, interval int) (bool, int, error) {
	seq, err := u.client.InsertedID(DefaultCollectionURLName)
	if err != nil {
		return false, 0, err
	}

	result, err := u.coll.InsertOne(
		ctx,
		bson.M{
			"id":       seq,
			"url":      url,
			"interval": interval,
		},
	)

	if err != nil {
		return false, 0, err
	}

	if result.InsertedID == nil {
		return false, 0, nil
	}

	return true, seq, nil
}

func (u *urlrepo) UpdateOneByID(ctx context.Context, id int, update interface{}) (bool, error) {
	result, err := u.coll.UpdateOne(
		ctx,
		bson.M{
			"id": id,
		}, bson.M{
			"$set": update,
		},
	)

	if err != nil {
		return false, err
	}

	return result.ModifiedCount > 0, nil
}

func (u *urlrepo) DeleteOneByID(ctx context.Context, id int) (bool, error) {
	result, err := u.coll.DeleteOne(
		ctx,
		bson.M{
			"id": id,
		},
	)
	if err != nil {
		return false, err
	}

	if result.DeletedCount == 0 {
		return false, nil
	}

	return true, nil
}

// TODO: tests
func (u *urlrepo) Scroll(ctx context.Context, f func([]metav1.URL)) error {
	cursor, err := u.coll.Find(ctx, bson.M{})
	if err != nil {
		return err
	}

	i := 1

	var urls []metav1.URL

	for cursor.Next(ctx) {
		var url metav1.URL

		if err := cursor.Decode(&url); err != nil {
			return err
		}

		urls = append(urls, url)

		if i >= u.defaultScroll {
			f(urls)
			urls = urls[:0]

			i = 1
		} else {
			i++
		}
	}

	f(urls)

	return nil
}
