/* Licensed to Elodina Inc. under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package zipkin

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/elodina/go-zipkin/log"
	producer "github.com/elodina/siesta-producer"
	"gopkg.in/spacemonkeygo/monitor.v1/trace/gen-go/zipkin"
)

type KafkaCollector struct {
	kafkaProducer *producer.KafkaProducer
	topic         string
}

func (c KafkaCollector) Collect(s *zipkin.Span) {
	t := thrift.NewTMemoryBuffer()
	p := thrift.NewTBinaryProtocolTransport(t)
	err := s.Write(p)
	if err != nil {
		log.Logger.Warn("Couldn't serialize span: ", err)
		return
	}

	// TODO: latter version of Siesta provides channel to hook up on which streams sending results, so need to use that.
	c.kafkaProducer.Send(&producer.ProducerRecord{Topic: c.topic, Value: t.Buffer.Bytes()})
}
