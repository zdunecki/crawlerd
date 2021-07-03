package pubsub

type PubSub interface {
	Publish(topic string, data []byte) error
	Subscribe(topic string, cb func([]byte)) error
}
