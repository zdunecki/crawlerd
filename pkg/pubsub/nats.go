package pubsub

import (
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type natsPubSub struct {
	client *nats.Conn
}

func NewNATSServer(configFile string) error {
	s, err := server.NewServer(&server.Options{
		ConfigFile: configFile,
	})
	if err != nil {
		return err
	}
	s.Start()
	return s.Reload()
}

//TODO: options
func NewNATS() (PubSub, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}

	return &natsPubSub{
		client: nc,
	}, nil
}

func (n *natsPubSub) Publish(subj string, data []byte) error {
	return n.client.Publish(subj, data)
}

func (n *natsPubSub) Subscribe(subj string, cb func([]byte)) error {
	var err error
	_, err = n.client.Subscribe(subj, func(msg *nats.Msg) {
		cb(msg.Data)
	})

	return err
}
