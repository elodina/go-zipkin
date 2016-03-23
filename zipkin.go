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
	"github.com/elodina/siesta"
	producer "github.com/elodina/siesta-producer"
	"github.com/spacemonkeygo/monotime"
	"golang.org/x/net/context"
	"gopkg.in/spacemonkeygo/monitor.v1/trace"
	"gopkg.in/spacemonkeygo/monitor.v1/trace/gen-go/zipkin"
)

type TraceConfig struct {
	ServiceName         string
	SampleRate          float64
	BrokerList          []string
	Topic               string
	Ip                  int32
	Port                int16
	KafkaProducerConfig *producer.ProducerConfig
}

type Tracer struct {
	manager   *trace.SpanManager
	collector *KafkaCollector
}

func NewTraceConfig(serviceName string, sampleRate float64, brokerList []string) *TraceConfig {
	producerConfig := producer.NewProducerConfig()
	producerConfig.BatchSize = 200
	producerConfig.ClientID = "zipkin"
	return &TraceConfig{
		serviceName,
		sampleRate,
		brokerList,
		"zipkin",
		127*256*256*256 + 1, // TODO: should find out actual local IP by default
		0,
		producerConfig,
	}
}

func NewTracer(config *TraceConfig) (*Tracer, error) {
	producerConfig := config.KafkaProducerConfig
	kafkaConnectorConfig := siesta.NewConnectorConfig()
	kafkaConnectorConfig.BrokerList = config.BrokerList
	connector, err := siesta.NewDefaultConnector(kafkaConnectorConfig)
	if err != nil {
		return nil, err
	}
	kafkaProducer := producer.NewKafkaProducer(producerConfig, producer.ByteSerializer, producer.ByteSerializer, connector)

	tracer := &Tracer{trace.NewSpanManager(), &KafkaCollector{kafkaProducer, config.Topic}}
	tracer.manager.Configure(config.SampleRate, false, &zipkin.Endpoint{
		Ipv4:        config.Ip,
		Port:        config.Port,
		ServiceName: config.ServiceName})
	tracer.manager.RegisterTraceCollector(tracer.collector)
	return tracer, nil
}

func (t *Tracer) InitServer(spanName string, r trace.Request) context.Context {
	ctx := t.createContextFromRequest(spanName, r)

	Annotate(ctx, zipkin.SERVER_RECV)

	return ctx
}

func (t *Tracer) TraceClient(ctx *context.Context, spanName string) func(*error) {
	return t.manager.TraceWithSpanNamed(ctx, spanName)
}

func (t *Tracer) TraceClientSent(ctx context.Context, spanName string) {
	t.manager.TraceWithSpanNamed(&ctx, spanName)
	t.CollectFromContext(ctx)
}

func Annotate(ctx context.Context, v string) {
	if span, ok := trace.SpanFromContext(ctx); ok {
		span.AnnotateTimestamp(v, monotime.Now(), nil, nil)
	}
}

func (t *Tracer) Collect(s *trace.Span) {
	if !s.TraceDisabled() {
		data := s.Export()
		t.collector.Collect(data)
	}
}

func (t *Tracer) CollectFromContext(ctx context.Context) {
	if span, ok := trace.SpanFromContext(ctx); ok {
		t.Collect(span)
	}
}

func (t *Tracer) createContextFromRequest(spanName string, r trace.Request) context.Context {
	span := t.manager.NewSpanFromRequest(spanName, r)
	return trace.ContextWithSpan(context.Background(), span)
}

func SampledSpanInfo(ctx context.Context) (s *trace.Span, ok bool) {
	currentSpan, ok := trace.SpanFromContext(ctx)
	return currentSpan, ok && !currentSpan.TraceDisabled()
}
