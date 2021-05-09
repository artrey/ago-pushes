package main

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

type CLI struct {
	producer sarama.SyncProducer
	topic    string
}

func NewCli(producer sarama.SyncProducer, topic string) *CLI {
	return &CLI{
		producer: producer,
		topic:    topic,
	}
}

var (
	app = kingpin.New("client", "A command-line client to send message into kafka.")

	userId  = app.Flag("userId", "Identifier of user").Required().Int64()
	message = app.Flag("message", "Message for pushing").Required().String()
)

type Message struct {
	UserID int64  `json:"userId"`
	Text   string `json:"text"`
}

func (cli *CLI) Run() error {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	data, err := json.Marshal(Message{
		UserID: *userId,
		Text:   *message,
	})
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: cli.topic,
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = cli.producer.SendMessage(msg)
	return err
}
