package zipkin

import (
	"github.com/elodina/siesta-producer"
	"github.com/yanzay/log"
)

type KafkaCollector struct {
	producer *producer.KafkaProducer
	topic string
}

func (kc *KafkaCollector) Collect(bytes []byte) {
	log.Debugf("[Zipkin] Collecting bytes: %v", bytes)
	kc.producer.Send(&producer.ProducerRecord{Topic: kc.topic, Value: bytes})
	log.Debugf("[Zipkin] Bytes collected")
}
