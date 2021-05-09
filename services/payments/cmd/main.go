package main

import (
	"errors"
	"github.com/Shopify/sarama"
	"log"
	"os"
	"time"
)

const (
	defaultBrokerURL  = "localhost:9093"
	defaultKafkaTopic = "tokens"
)

func main() {
	brokerURL, ok := os.LookupEnv("APP_BROKER_URL")
	if !ok {
		brokerURL = defaultBrokerURL
	}

	topic, ok := os.LookupEnv("APP_KAFKA_TOPIC")
	if !ok {
		topic = defaultKafkaTopic
	}

	if err := execute(brokerURL, topic); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(brokerURL string, topic string) error {
	sarama.Logger = log.New(os.Stdout, "", log.Ltime)
	config := sarama.NewConfig()
	config.ClientID = "payments"
	config.Producer.Return.Successes = true
	config.Version = sarama.V2_6_0_0

	producer, err := waitForKafka(brokerURL, config)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := producer.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(cerr)
		}
	}()

	cli := NewCli(producer, topic)
	return cli.Run()
}

func waitForKafka(brokerURL string, config *sarama.Config) (sarama.SyncProducer, error) {
	for {
		select {
		case <-time.After(time.Minute):
			return nil, errors.New("can't connect to kafka")
		default:
		}

		producer, err := sarama.NewSyncProducer([]string{brokerURL}, config)
		if err != nil {
			log.Print(err)
			time.Sleep(time.Second)
			continue
		}

		return producer, nil
	}
}
