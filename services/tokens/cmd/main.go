package main

import (
	"context"
	"errors"
	"github.com/Shopify/sarama"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net"
	"net/http"
	"os"
	"time"
	"tokens/cmd/app"
	"tokens/pkg/tokens"
)

const (
	defaultPort       = "9999"
	defaultHost       = "0.0.0.0"
	defaultDSN        = "postgres://app:pass@localhost:5532/db"
	defaultBrokerURL  = "localhost:9093"
	defaultKafkaTopic = "tokens"
	defaultKafkaGroup = "tokens-core"
)

func main() {
	port, ok := os.LookupEnv("APP_PORT")
	if !ok {
		port = defaultPort
	}

	host, ok := os.LookupEnv("APP_HOST")
	if !ok {
		host = defaultHost
	}

	dsn, ok := os.LookupEnv("APP_DSN")
	if !ok {
		dsn = defaultDSN
	}

	brokerURL, ok := os.LookupEnv("APP_BROKER_URL")
	if !ok {
		brokerURL = defaultBrokerURL
	}

	kafkaTopic, ok := os.LookupEnv("APP_KAFKA_TOPIC")
	if !ok {
		kafkaTopic = defaultKafkaTopic
	}

	kafkaGroup, ok := os.LookupEnv("APP_KAFKA_GROUP")
	if !ok {
		kafkaGroup = defaultKafkaGroup
	}

	if err := execute(net.JoinHostPort(host, port), dsn, brokerURL, kafkaTopic, kafkaGroup); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(addr string, dsn string, brokerURL string, transactionsTopic string, group string) error {
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		log.Print(err)
		return err
	}
	defer pool.Close()

	tokensSvc := tokens.NewService(pool)
	mux := chi.NewRouter()

	application := app.NewServer(tokensSvc, mux)
	err = application.Init()
	if err != nil {
		log.Print(err)
		return err
	}
	server := &http.Server{
		Addr:    addr,
		Handler: application,
	}

	sarama.Logger = log.New(os.Stdout, "", log.Ltime)
	config := sarama.NewConfig()
	config.ClientID = "tokens"
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false
	config.Version = sarama.V2_6_0_0

	consumerGroup, err := waitForKafka(brokerURL, group, config)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := consumerGroup.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(cerr)
		}
	}()

	subscriber := app.NewSubscriber(tokensSvc, consumerGroup)

	errs := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			errs <- err
		}
	}()
	go func() {
		for {
			err := consumerGroup.Consume(context.Background(), []string{transactionsTopic}, subscriber)
			if err != nil {
				errs <- err
				return
			}
		}
	}()

	return <-errs
}

func waitForKafka(brokerURL string, group string, config *sarama.Config) (sarama.ConsumerGroup, error) {
	for {
		select {
		case <-time.After(time.Minute):
			return nil, errors.New("can't connect to kafka")
		default:
		}

		consumerGroup, err := sarama.NewConsumerGroup([]string{brokerURL}, group, config)
		if err != nil {
			log.Print(err)
			time.Sleep(time.Second)
			continue
		}

		return consumerGroup, nil
	}
}
