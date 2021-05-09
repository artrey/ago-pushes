package app

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"log"
	"tokens/cmd/app/dto"
	"tokens/pkg/tokens"
)

type Subscriber struct {
	tokensSvc     *tokens.Service
	consumerGroup sarama.ConsumerGroup
}

func NewSubscriber(tokensSvc *tokens.Service, consumerGroup sarama.ConsumerGroup) *Subscriber {
	return &Subscriber{tokensSvc: tokensSvc, consumerGroup: consumerGroup}
}

func (s *Subscriber) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (s *Subscriber) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (s *Subscriber) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("received message from topic: %s %s", claim.Topic(), message.Value)
		session.MarkMessage(message, "")
		session.Commit()

		var pushMessage dto.Message
		err := json.Unmarshal(message.Value, &pushMessage)
		if err != nil {
			log.Print(err)
			continue
		}

		token, err := s.tokensSvc.GetToken(context.Background(), pushMessage.UserID)
		if err != nil {
			log.Print(err)
			continue
		}

		// TODO: send via Firebase
		log.Printf("send notification to mobile app with token: %s and message: %s",
			token.PushToken, pushMessage.Text)
	}

	return nil
}
