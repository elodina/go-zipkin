package ztrace

import (
	"gopkg.in/spacemonkeygo/monitor.v1/trace"
	"gopkg.in/spacemonkeygo/monitor.v1/trace/gen-go/zipkin"
	"github.com/spacemonkeygo/monotime"
	"golang.org/x/net/context"
	"github.com/stealthly/siesta"
	"time"
)

type TraceConfig struct {
	ServiceName string
	SampleRate  float64
	BrokerList  []string
	Topic       string
	Ip          int32
	Port        int16
	KafkaProducerConfig *siesta.ProducerConfig
}

func NewTraceConfig(serviceName string, sampleRate float64, brokerList []string) (*TraceConfig) {
	return &TraceConfig{
		serviceName,
		sampleRate,
		brokerList,
		"zipkin",
		127 * 256 * 256 * 256 + 1, // TODO: should find out actual local IP by default
		0,
		&siesta.ProducerConfig{
			BatchSize:       200,
			ClientID:        "ztrace",
			MaxRequests:     10,
			SendRoutines:    10,
			ReceiveRoutines: 10,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			RequiredAcks:    1,
			AckTimeoutMs:    2000,
			Linger:          5 * time.Second,
		},
	}
}

func Initialize(config *TraceConfig) {

	producerConfig := config.KafkaProducerConfig
	kafkaConnectorConfig := siesta.NewConnectorConfig()
	kafkaConnectorConfig.BrokerList = config.BrokerList
	siesta.NewDefaultConnector(kafkaConnectorConfig)
	connector, err := siesta.NewDefaultConnector(kafkaConnectorConfig)
	if err != nil {
		panic(err)
	}
	producer := siesta.NewKafkaProducer(producerConfig, siesta.ByteSerializer, siesta.ByteSerializer, connector)
	collector = KafkaCollector{*producer, config.Topic}

	trace.Configure(config.SampleRate, false, &zipkin.Endpoint{
		Ipv4:        config.Ip,
		Port:        config.Port,
		ServiceName: config.ServiceName})
	trace.RegisterTraceCollector(collector)
}

func Collect(s *trace.Span) {
	if (!s.TraceDisabled()) {
		data := s.Export()
		collector.Collect(data)
	}
}

func CreateContextFromRequest(spanName string, r trace.Request) (context.Context) {
	span := trace.NewSpanFromRequest(spanName, r)
	return trace.ContextWithSpan(context.Background(), span)
}

func CollectFromContext(ctx context.Context) {
	if span, ok := trace.SpanFromContext(ctx); ok {
		Collect(span)
	}
}

func Annotate(ctx context.Context, v string) {
	if span, ok := trace.SpanFromContext(ctx); ok {
		span.AnnotateTimestamp(v, monotime.Now(), nil, nil)
	}
}

func InitServer(spanName string, r trace.Request) (context.Context) {
	ctx := CreateContextFromRequest(spanName, r)

	Annotate(ctx, zipkin.SERVER_RECV)

	return ctx
}

func TraceClient(ctx *context.Context, spanName string) func(*error) {
	return trace.TraceWithSpanNamed(ctx, spanName)
}

func TraceClientSent(ctx context.Context, spanName string) {
	trace.TraceWithSpanNamed(&ctx, spanName)
	CollectFromContext(ctx)
}

var (
	collector KafkaCollector
)