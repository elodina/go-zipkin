package zipkin

import (
	"math/rand"
	"sync"
	"time"

	"github.com/elodina/go-zipkin/gen-go/zipkincore"
	"github.com/elodina/siesta-producer"
	"github.com/yanzay/log"
)

var localhost int32 = 127*256*256*256 + 1

type Collector interface {
	Collect([]byte)
}

type Tracer struct {
	collector Collector
	host      string
	name      string
	rate      int
}

func NewTracer(name string, host string, rate int, producer *producer.KafkaProducer) *Tracer {
	log.Infof("[Zipkin] Creating new tracer for service %s with rate 1:%d", name, rate)
	collector := &KafkaCollector{producer: producer}
	tracer := &Tracer{host: host, collector: collector, rate: rate, name: name}
	return tracer
}

func (t *Tracer) NewSpan(name string) *Span {
	log.Infof("[Zipkin] Creating new span: %s", name)
	sampled := rand.Intn(t.rate) == 0
	if !sampled {
		return &Span{sampled: false}
	}
	span := newSpan(name, newID(), nil, t.name)
	span.collector = t.collector
	span.sampled = true
	return span
}

type Span struct {
	sync.Mutex
	span        *zipkincore.Span
	collector   Collector
	children    []*Span
	sampled     bool
	serviceName string
}

func newSpan(name string, traceID int64, parentID *int64, serviceName string) *Span {
	zipkinSpan := &zipkincore.Span{
		Name:              name,
		ID:                newID(),
		TraceID:           traceID,
		ParentID:          parentID,
		Annotations:       make([]*zipkincore.Annotation, 0),
		BinaryAnnotations: make([]*zipkincore.BinaryAnnotation, 0),
	}

	return &Span{span: zipkinSpan, serviceName: serviceName}
}

func (s *Span) Sampled() bool {
	return s.sampled
}

func (s *Span) TraceID() int64 {
	return s.span.TraceID
}

func (s *Span) ParentID() *int64 {
	return s.span.ParentID
}

func (s *Span) ID() int64 {
	return s.span.ID
}

func (s *Span) ServerReceive() {
	log.Infof("[Zipkin] ServerReceive")
	s.Annotate(zipkincore.SERVER_RECV)
}

func (s *Span) ServerSend() {
	log.Infof("[Zipkin] ServerSend")
	s.Annotate(zipkincore.SERVER_SEND)
}

func (s *Span) ServerSendAndCollect() {
	s.ServerSend()
	s.send()
}

func (s *Span) ClientSend() {
	log.Infof("[Zipkin] ClientSend")
	s.Annotate(zipkincore.CLIENT_SEND)
}

func (s *Span) ClientReceive() {
	log.Infof("[Zipkin] ClientReceive")
	s.Annotate(zipkincore.CLIENT_RECV)
}

func (s *Span) ClientReceiveAndCollect() {
	s.ClientReceive()
	s.send()
}

func (s *Span) NewChild(name string) *Span {
	log.Infof("[Zipkin] Creating new child span: %s", name)
	if !s.sampled {
		return &Span{}
	}
	child := newSpan(name, s.span.TraceID, &s.span.ID, s.serviceName)
	child.collector = s.collector
	s.Lock()
	s.children = append(s.children, child)
	s.Unlock()
	return child
}

func (s *Span) send() error {
	log.Infof("[Zipkin] Sending spans: %b", s.sampled)
	if !s.sampled {
		return nil
	}
	log.Infof("[Zipkin] Serializing span: %v", s.span)
	bytes, err := SerializeSpan(s.span)
	if err != nil {
		return err
	}
	s.collector.Collect(bytes)
	return nil
}

func (s *Span) Annotate(value string) {
	if !s.sampled {
		return
	}
	now := nowMicrosecond()
	annotation := &zipkincore.Annotation{
		Value:     value,
		Timestamp: *now,
		Host: &zipkincore.Endpoint{
			ServiceName: s.serviceName,
			Ipv4:        localhost,
			Port:        0,
		},
	}
	s.Lock()
	s.span.Annotations = append(s.span.Annotations, annotation)
	s.Unlock()
}

func nowMicrosecond() *int64 {
	now := time.Now().UnixNano() / 1000
	return &now
}

func newID() int64 {
	return rand.Int63()
}
