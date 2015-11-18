package main

import (
	"golang.org/x/net/context"
	"gopkg.in/spacemonkeygo/monitor.v1/trace"
	"github.com/elodina/go-zipkin"
	"time"
	"github.com/stealthly/siesta"
	"fmt"
)

func main() {

	// First, we initialize our stuff in main
	traceConfig := ztrace.NewTraceConfig("some_service", 0.5, []string{"slave0:31001"})

	// We want kind of a simpler Kafka default config for a sample application
	traceConfig.KafkaProducerConfig = &siesta.ProducerConfig{
		BatchSize:       1,
		ClientID:        "ztrace",
		MaxRequests:     10,
		SendRoutines:    10,
		ReceiveRoutines: 10,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		RequiredAcks:    1,
		AckTimeoutMs:    2000,
		Linger:          1 * time.Second,
	}

	ztrace.Initialize(traceConfig)

	// Suppose we got some event, and start tracing from there.
	// We should get ~5 traces in a span storage with sample rate set to 0.5
	for j := 0; j <= 9; j++ {
		onEvent(j)
	}

	// Just giving some time for a Kafka producer to grab and send messages further
	time.Sleep(100 * time.Millisecond)
}

func onEvent(j int) {

	// Need to create new context based on a new span
	// In case if we start with tracing, we simply call Context.Background() instead
	ctx := ztrace.InitServer(fmt.Sprint("incoming_call_", j), trace.Request{})

	// Presume, we do an RPC call
	sendRPC(ctx)

	// And fire and forget after that
	sendFireAndForget(ctx)

	// We may want to submit some custom annotations along with the current span
	ztrace.Annotate(ctx, "Completed Processing")

	// Will also need to explicitly send span to collector
	ztrace.CollectFromContext(ctx)
}

func sendRPC(ctx context.Context) {
	// This adds client send annotation right away, and adds client receive annotation on defer
	defer ztrace.TraceClient(&ctx, "rpc_call")(nil)

	time.Sleep(200 * time.Millisecond)
}

func sendFireAndForget(ctx context.Context) {
	//This only adds client send annotation right away
	ztrace.TraceClientSent(&ctx, "faf_call")

	time.Sleep(100 * time.Millisecond)
}