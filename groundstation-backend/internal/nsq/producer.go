package nsq

import (
	"encoding/json"
	"groundstation-backend/internal/config"
	"sync"

	"github.com/nsqio/go-nsq"
)

var producer *nsq.Producer
var producerOnce sync.Once
var producerErr error

func InitProducer() error {
	producerOnce.Do(func() {
		cfg := config.AppConfig.NSQ
		config := nsq.NewConfig()
		producer, producerErr = nsq.NewProducer(cfg.NSQDAddress, config)
		if producerErr != nil {
			return
		}
		producerErr = producer.Ping()
	})
	return producerErr
}

func GetProducer() (*nsq.Producer, error) {
	if err := InitProducer(); err != nil {
		return nil, err
	}
	return producer, nil
}

func Publish(topic string, message interface{}) error {
	producer, err := GetProducer()
	if err != nil {
		return err
	}

	var data []byte
	switch v := message.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		data, err = json.Marshal(message)
		if err != nil {
			return err
		}
	}

	return producer.Publish(topic, data)
}

func PublishDeferred(topic string, message interface{}, delay_ms int) error {
	producer, err := GetProducer()
	if err != nil {
		return err
	}

	var data []byte
	switch v := message.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		data, err = json.Marshal(message)
		if err != nil {
			return err
		}
	}

	return producer.DeferredPublish(topic, int(delay_ms), data)
}

func PublishMultiple(topic string, messages []interface{}) error {
	producer, err := GetProducer()
	if err != nil {
		return err
	}

	data := make([][]byte, len(messages))
	for i, msg := range messages {
		switch v := msg.(type) {
		case []byte:
			data[i] = v
		case string:
			data[i] = []byte(v)
		default:
			jsonData, err := json.Marshal(msg)
			if err != nil {
				return err
			}
			data[i] = jsonData
		}
	}

	return producer.MultiPublish(topic, data)
}

func StopProducer() {
	if producer != nil {
		producer.Stop()
	}
}
