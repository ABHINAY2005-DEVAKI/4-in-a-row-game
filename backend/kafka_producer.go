package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Shopify/sarama"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(brokers []string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 5 * time.Second
	p, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &KafkaProducer{producer: p}, nil
}

func (kp *KafkaProducer) SendEvent(topic string, payload any) {
	if kp == nil {
		return
	}
	b, _ := json.Marshal(payload)
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(b),
	}
	_, _, err := kp.producer.SendMessage(msg)
	if err != nil {
		log.Println("kafka send error:", err)
	}
}
