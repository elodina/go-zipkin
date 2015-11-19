package main

import (
	"golang.org/x/net/context"
	"gopkg.in/spacemonkeygo/monitor.v1/trace"
	"github.com/elodina/go-zipkin"
	"time"
	"github.com/stealthly/siesta"
	"fmt"
	z "gopkg.in/spacemonkeygo/monitor.v1/trace/gen-go/zipkin"
)

var (
	samplerTracer *zipkin.Tracer
	fractionalTracer *zipkin.Tracer
)


func main() {

	// We want kind of a simpler Kafka default config for a sample application
	kafkaProducerConfig := &siesta.ProducerConfig{
		BatchSize:       1,
		ClientID:        "go-zipkin",
		MaxRequests:     10,
		SendRoutines:    10,
		ReceiveRoutines: 10,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		RequiredAcks:    1,
		AckTimeoutMs:    2000,
		Linger:          1 * time.Second,
	}

	// For most cases you'll need only one tracer per app, but we'll showcase how to create several of them.
	samplerTraceConfig := zipkin.NewTraceConfig("some_service", 1.0, []string{"slave0:31001"})
	samplerTraceConfig.KafkaProducerConfig = kafkaProducerConfig

	fractionalTraceConfig := zipkin.NewTraceConfig("some_service", 0.5, []string{"slave0:31001"})
	fractionalTraceConfig.KafkaProducerConfig = kafkaProducerConfig

	samplerTracer = initTracer(samplerTraceConfig)

	fractionalTracer = initTracer(fractionalTraceConfig)

	//Let's call a function, which will use samplerTracer, so it will trace all the requests
	onCriticalEvent()

	// Here we should get ~5 traces in a span storage, as the sample rate set to 0.5
	for j := 0; j <= 9; j++ {
	  onEvent(j)
	}

	// Just giving some time for a Kafka producer to grab and send messages further
	time.Sleep(50 * time.Millisecond)
}

func initTracer(config *zipkin.TraceConfig) *zipkin.Tracer {
	tracer,err := zipkin.NewTracer(config)
	if (err != nil) {
		panic(err)
	}
	return tracer
}

func onCriticalEvent() {
	ctx := samplerTracer.InitServer(fmt.Sprint("critical_call"), trace.Request{})

	time.Sleep(50 * time.Millisecond)

	zipkin.Annotate(ctx, z.SERVER_SEND)
	samplerTracer.CollectFromContext(ctx)
}

func onEvent(j int) {

	// Need to create new context based on a new span
	// In case if we start with tracing, we simply call Context.Background() instead
	ctx := fractionalTracer.InitServer(fmt.Sprint("incoming_call_", j), trace.Request{})

	// Presume, we do an RPC call
	sendRPC(ctx)

	// And fire and forget after that
	sendFireAndForget(ctx)

	// We may want to submit some custom annotations along with the current span
	zipkin.Annotate(ctx, "Completed Processing")

	// Will also need to explicitly send span to collector
	fractionalTracer.CollectFromContext(ctx)
}

func sendRPC(ctx context.Context) {
	// This adds client send annotation right away, and adds client receive annotation on defer
	defer fractionalTracer.TraceClient(ctx, "rpc_call")(nil)

	time.Sleep(50 * time.Millisecond)
}

func sendFireAndForget(ctx context.Context) {
	//This only adds client send annotation right away
	fractionalTracer.TraceClientSent(ctx, "faf_call")

	time.Sleep(50 * time.Millisecond)
}