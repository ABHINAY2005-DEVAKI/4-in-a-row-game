package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/Shopify/sarama"
)

type Consumer struct{}

func (c *Consumer) Setup(s sarama.ConsumerGroupSession) error   { return nil }
func (c *Consumer) Cleanup(s sarama.ConsumerGroupSession) error { return nil }
func (c *Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var payload map[string]interface{}
		json.Unmarshal(msg.Value, &payload)
		fmt.Println("analytics event:", string(msg.Value))
		sess.MarkMessage(msg, "")
	}
	return nil
}

func main() {
	brokers := []string{"localhost:9092"}
	group := "analytics-group"
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	client, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		log.Fatal(err)
	}
	consumer := &Consumer{}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			if err := client.Consume(ctx, []string{"game_start", "move", "game_end"}, consumer); err != nil {
				log.Println("error:", err)
			}
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	cancel()
	client.Close()
}


