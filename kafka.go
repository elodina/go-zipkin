package ztrace

import (
	"github.com/stealthly/siesta"
	"gopkg.in/spacemonkeygo/monitor.v1/trace/gen-go/zipkin"
	"git.apache.org/thrift.git/lib/go/thrift"
	"time"
	"fmt"
)

type KafkaCollector struct {
	producer siesta.KafkaProducer
	topic    string
}

func (c KafkaCollector) Collect(s *zipkin.Span) {
	t := thrift.NewTMemoryBuffer()
	p := thrift.NewTBinaryProtocolTransport(t)
	err := s.Write(p)
	if err != nil {
		fmt.Println("error occurred: ", err)
		return
	}

	println("sending traces")
	recordMetadata := c.producer.Send(&siesta.ProducerRecord{Topic: c.topic, Value: t.Buffer.Bytes()})

	select {
	case x := <-recordMetadata:
		if (x.Error != siesta.ErrNoError) {
			fmt.Println("error occurred: ", x.Error)
		}
		println("sent span")
	case <-time.After(5 * time.Second):
	    // TODO: not sure on timeout duration?
		println("timeout exceeded")
	}

}