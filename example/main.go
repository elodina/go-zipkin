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

package main

import (
	"fmt"
	"github.com/elodina/go-zipkin"
	producer "github.com/elodina/siesta-producer"
	"golang.org/x/net/context"
	"gopkg.in/spacemonkeygo/monitor.v1/trace"
	z "gopkg.in/spacemonkeygo/monitor.v1/trace/gen-go/zipkin"
	"time"
)

var (
	samplerTracer    *zipkin.Tracer
	fractionalTracer *zipkin.Tracer
)

func main() {

	// We want kind of a simpler Kafka default config for a sample application
	kafkaProducerConfig := producer.NewProducerConfig()
	kafkaProducerConfig.BatchSize = 1
	kafkaProducerConfig.ClientID = "go-zipkin"

	// For most cases you'll need only one tracer per app, but we'll showcase how to create several of them.
	samplerTraceConfig := zipkin.NewTraceConfig("some_service", 1.0, []string{"broker-70.service.cluster2:31250"})
	samplerTraceConfig.KafkaProducerConfig = kafkaProducerConfig

	fractionalTraceConfig := zipkin.NewTraceConfig("some_service", 0.5, []string{"broker-70.service.cluster2:31250"})
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
	tracer, err := zipkin.NewTracer(config)
	if err != nil {
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
	defer fractionalTracer.TraceClient(&ctx, "rpc_call")(nil)

	if spanInfo, ok := zipkin.SampledSpanInfo(ctx); ok {
		fmt.Print("Span info to send:")
		fmt.Println(spanInfo)
	}
	time.Sleep(50 * time.Millisecond)
}

func sendFireAndForget(ctx context.Context) {
	//This only adds client send annotation right away
	fractionalTracer.TraceClientSent(ctx, "faf_call")

	time.Sleep(50 * time.Millisecond)
}
