package pubsub

import (
	"context"

	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

type kafkapusub struct {
	broker         string
	partitionConns map[*kafka.Partition]*kafka.Conn
	conn           *kafka.Conn
}

// TODO: options
func NewKafka(broker string) (PubSub, error) {
	conn, err := kafka.DialContext(context.Background(), "tcp", broker)
	if err != nil {
		return nil, err
	}
	return &kafkapusub{
		conn: conn,
	}, nil
}

func (k *kafkapusub) Publish(topic string, data []byte) error {
	writeMessage := func(conn *kafka.Conn) error {
		_, err := conn.WriteMessages(kafka.Message{
			Value: data,
		})

		return err
	}

	if k.partitionConns != nil && len(k.partitionConns) > 0 {
		for _, conn := range k.partitionConns {
			if err := writeMessage(conn); err != nil {
				log.Error(err)
			}
		}

		return nil
	}

	partitions, err := k.conn.ReadPartitions(topic)
	if err != nil {
		return err
	}

	for _, partition := range partitions {
		conn, err := kafka.DialPartition(context.Background(), "tcp", k.broker, partition)
		if err != nil {
			log.Error(err)
			continue
		}
		k.partitionConns[&partition] = conn

		if err := writeMessage(conn); err != nil {
			log.Error(err)
		}
	}

	return nil
}

// TODO:
func (k *kafkapusub) Subscribe(topic string, cb func([]byte)) error {
	panic("implement me")
}
