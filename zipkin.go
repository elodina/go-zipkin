package zipkin

import (
	"math/rand"
	"sync"
	"time"

	"github.com/elodina/go-zipkin/gen-go/zipkincore"
	"github.com/elodina/siesta-producer"
	"github.com/elodina/siesta"
	"github.com/yanzay/log"
	"net"
	"errors"
	"encoding/binary"
	"bytes"
	"github.com/elodina/go-avro"
)

var localhost int32 = 127 * 256 * 256 * 256 + 1

type Collector interface {
	Collect([]byte)
}

type Tracer struct {
	collector   Collector
	ip          int32
	port        int16
	serviceName string
	rate        int
}

func NewTracer(serviceName string, rate int, producer *producer.KafkaProducer, ip string, port int16, topic string) *Tracer {
	log.Infof("[Zipkin] Creating new tracer for service %s with rate 1:%d and topic %s", serviceName, rate, topic)
	collector := &KafkaCollector{producer: producer, topic: topic}

	convertedIp, err := convertIp(ip); if (err != nil) {
		convertedIp = &localhost
	}

	tracer := &Tracer{ip: *convertedIp, port: port, collector: collector, rate: rate, serviceName: serviceName}
	return tracer
}

func convertIp(ip string) (*int32, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, errors.New("Unable to parse given ip")
	}

	var ipInInt int32
	buf := bytes.NewReader([]byte{parsedIP[0], parsedIP[1], parsedIP[2], parsedIP[3]})
	err := binary.Read(buf, binary.LittleEndian, &ipInInt); if err == nil {
		return &ipInInt, nil
	} else {
		return nil, err
	}
}

func DefaultTopic() string {
	return "zipkin"
}

func DefaultPort() int16 {
	return 0
}

func DefaultProducer(brokerList []string) (*producer.KafkaProducer, error) {
	producerConfig := producer.NewProducerConfig()
	producerConfig.BatchSize = 200
	producerConfig.ClientID = "zipkin"
	kafkaConnectorConfig := siesta.NewConnectorConfig()
	kafkaConnectorConfig.BrokerList = brokerList
	connector, err := siesta.NewDefaultConnector(kafkaConnectorConfig)
	if err != nil {
		return nil, err
	}
	return producer.NewKafkaProducer(producerConfig, producer.ByteSerializer, producer.ByteSerializer, connector), nil
}

func LocalNetworkIP() string {
	ip, err := determineLocalIp()
	if err != nil {
		return ip
	} else {
		log.Infof("Unable to determine local network IP address, going with localhost IP")
		return "127.0.0.1"
	}
}

func determineLocalIp() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags & net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags & net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("Not connected to network")
}

func (t *Tracer) NewSpan(name string) *Span {
	log.Debugf("[Zipkin] Creating new span: %s", name)
	sampled := rand.Intn(t.rate) == 0
	if !sampled {
		return &Span{sampled: false}
	}
	span := newSpan(name, newID(), newID(), nil, t.serviceName)
	span.collector = t.collector
	span.port = t.port
	span.ip = t.ip
	span.sampled = true
	return span
}

type Span struct {
	sync.Mutex
	span        *zipkincore.Span
	collector   Collector
	ip          int32
	port        int16
	sampled     bool
	serviceName string
}

func newSpan(name string, traceID int64, spanId int64, parentID *int64, serviceName string) *Span {
	zipkinSpan := &zipkincore.Span{
		Name:              name,
		ID:                spanId,
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
	log.Debugf("[Zipkin] ServerReceive")
	s.Annotate(zipkincore.SERVER_RECV)
}

func (s *Span) ServerSend() {
	log.Debugf("[Zipkin] ServerSend")
	s.Annotate(zipkincore.SERVER_SEND)
}

func (s *Span) ServerSendAndCollect() {
	s.ServerSend()
	s.Collect()
}

func (s *Span) ServerReceiveAndCollect() {
	s.ServerReceive()
	s.Collect()
}

func (s *Span) ClientSend() {
	log.Debugf("[Zipkin] ClientSend")
	s.Annotate(zipkincore.CLIENT_SEND)
}

func (s *Span) ClientReceive() {
	log.Debugf("[Zipkin] ClientReceive")
	s.Annotate(zipkincore.CLIENT_RECV)
}

func (s *Span) ClientReceiveAndCollect() {
	s.ClientReceive()
	s.Collect()
}

func (s *Span) ClientSendAndCollect() {
	s.ClientSend()
	s.Collect()
}

func (s *Span) NewChild(name string) *Span {
	log.Debugf("[Zipkin] Creating new child span: %s", name)
	if !s.sampled {
		return &Span{}
	}
	child := newSpan(name, s.span.TraceID, newID(), &s.span.ID, s.serviceName)
	child.collector = s.collector
	child.ip = s.ip
	child.port = s.port
	child.sampled = true
	return child
}

func (t *Tracer) NewSpanFromAvro(name string, traceInfo interface{}) *Span {

	if traceInfo == nil {
		return t.NewSpanFromRequest(name, nil, nil, nil, nil);
	}

	traceInfoDef := traceInfo.(*avro.GenericRecord)

	var traceId *int64 = nil
	if traceIdAvro := traceInfoDef.Get("traceId"); traceIdAvro != nil {
		traceIdDef := traceIdAvro.(int64)
		traceId = &traceIdDef
	}

	var spanId *int64 = nil
	if spanIdAvro := traceInfoDef.Get("spanId"); spanIdAvro != nil {
		spanIdDef := spanIdAvro.(int64)
		spanId = &spanIdDef
	}

	var sampled *bool = nil
	if sampledAvro := traceInfoDef.Get("sampled"); sampledAvro != nil {
		sampledDef := sampledAvro.(bool)
		sampled = &sampledDef
	}

	var parentId *int64 = nil
	if parentIdAvro := traceInfoDef.Get("parentSpanId"); parentIdAvro != nil {
		parentIdDef := parentIdAvro.(int64)
		parentId = &parentIdDef
	}

	return t.NewSpanFromRequest(name, traceId, spanId, parentId, sampled)
}

func (t *Tracer) NewSpanFromRequest(name string, traceId *int64, spanId *int64, parentId *int64, sampled *bool) *Span {
	nonSampled := &Span{sampled: false}
	if sampled == nil {
		log.Debugf("[Zipkin] Empty trace info provided. Ignoring")
		return nonSampled
	}
	if !*sampled {
		log.Debugf("[Zipkin] The input trace info not sampled. Ignoring")
		return nonSampled
	}
	if spanId == nil || traceId == nil {
		log.Debugf("[Zipkin] The input trace info incomplete. Ignoring")
		return nonSampled
	}

	log.Debugf("[Zipkin] Creating new span %s from request: traceID %s, spanID %s, parentID %s, sampled %s", name,
	traceId, spanId, parentId, sampled)
	span := newSpan(name, *traceId, *spanId, nil, t.serviceName)
	span.collector = t.collector
	span.port = t.port
	span.ip = t.ip
	span.sampled = true
	return span
}

func (s *Span) GetAvroTraceInfo() *avro.GenericRecord {
	if s.sampled {
		traceInfo := avro.NewGenericRecord(NewTraceInfo().Schema())
		traceInfo.Set("traceId", s.TraceID())
		traceInfo.Set("spanId", s.ID())
		if parentId := s.ParentID(); parentId != nil {
			traceInfo.Set("parentSpanId", *parentId)
		}
		traceInfo.Set("sampled", true)
		return traceInfo
	} else {
		return nil
	}
}

func (s *Span) Collect() error {
	log.Debugf("[Zipkin] Sending spans: %b", s.sampled)
	if !s.sampled {
		return nil
	}
	log.Debugf("[Zipkin] Serializing span: %v", s.span)
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
			Ipv4:        s.ip,
			Port:        s.port,
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
