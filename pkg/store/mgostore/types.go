package mgostore

type Sequence struct {
	ObjectID string `bson:"_id"`
	ID       int    `bson:"id"`
}
