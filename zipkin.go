package zipkin

import (
	"github.com/elodina/go-zipkin/Godeps/_workspace/src/github.com/spacemonkeygo/monotime"
	"github.com/elodina/go-zipkin/Godeps/_workspace/src/github.com/stealthly/siesta"
	"github.com/elodina/go-zipkin/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/elodina/go-zipkin/Godeps/_workspace/src/gopkg.in/spacemonkeygo/monitor.v1/trace"
	"github.com/elodina/go-zipkin/Godeps/_workspace/src/gopkg.in/spacemonkeygo/monitor.v1/trace/gen-go/zipkin"
	"time"
)

type TraceConfig struct {
	ServiceName         string
	SampleRate          float64
	BrokerList          []string
	Topic               string
	Ip                  int32
	Port                int16
	KafkaProducerConfig *siesta.ProducerConfig
}

type Tracer struct {
	manager   *trace.SpanManager
	collector *KafkaCollector
}

func NewTraceConfig(serviceName string, sampleRate float64, brokerList []string) *TraceConfig {
	return &TraceConfig{
		serviceName,
		sampleRate,
		brokerList,
		"zipkin",
		127*256*256*256 + 1, // TODO: should find out actual local IP by default
		0,
		&siesta.ProducerConfig{
			BatchSize:       200,
			ClientID:        "go-zipkin",
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

func NewTracer(config *TraceConfig) (*Tracer, error) {
	producerConfig := config.KafkaProducerConfig
	kafkaConnectorConfig := siesta.NewConnectorConfig()
	kafkaConnectorConfig.BrokerList = config.BrokerList
	siesta.NewDefaultConnector(kafkaConnectorConfig)
	connector, err := siesta.NewDefaultConnector(kafkaConnectorConfig)
	if err != nil {
		return nil, err
	}
	producer := siesta.NewKafkaProducer(producerConfig, siesta.ByteSerializer, siesta.ByteSerializer, connector)

	tracer := &Tracer{trace.NewSpanManager(), &KafkaCollector{producer, config.Topic}}
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

func (t *Tracer) TraceClient(ctx context.Context, spanName string) func(*error) {
	return t.manager.TraceWithSpanNamed(&ctx, spanName)
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
