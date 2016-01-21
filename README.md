# Zipkin library for Go

Initial implementation of Zipkin tracing library for Go. Specifically created for collecting Zipkin spans using 
Kafka as a transport. Based on Spacemonkey Monitor: https://github.com/spacemonkeygo/monitor. Relies on using Google 
Context objects: http://blog.golang.org/context.

This library enables the following features:

 - Kafka span collector
 - Ability to collect spans to the collector whenever needed
 
## Quickstart
 
Install go-zipkin library and its dependencies. Before installation make sure to have your `GOROOT` and `GOPATH` 
installed properly. It's also necessary to have [Godep](https://github.com/tools/godep) installed and `PATH` variable 
containing your `GOPATH/bin` 

```
go get github.com/elodina/go-zipkin
cd $GOPATH/src/github.com/elodina/go-zipkin
godep restore
```

Import go-zipkin lib into your app and configure the tracer. Tracer is an instance of struct `zipkin.Tracer` which one 
should use for most of Zipkin-related interactions along the program lifecycle. For the most basic set of parameters 
one should specify name of the traced service, sample rate, and the set of Kafka broker endpoints.
 
**NOTE:** sample rate stands for how often Zipkin traces would be actually traced. For this parameter one should use a 
fraction number from 0 to 1, inclusive. Where 0 stands for no tracing, 1 - trace everything, etc.

```
import "github.com/elodina/go-zipkin"

// Created tracer to trace "service_name" service with sample rate 0.5 through "slave0:31001" Kafka broker endpoint.
traceConfig := zipkin.NewTraceConfig("service_name", 0.5, []string{"slave0:31001"})
tracer, err := zipkin.NewTracer(traceConfig)
```

After the tracer is created, you may use it for annotating and collecting spans. In case if you start your trace from 
scratch starting from "server received" event:

```
// This creates (but not yet collects!) "server received" annotation for a span with name "incoming_call" with no initial trace info.
// Note that in the second parameter you may pass the trace info coming from the client side, such as trace id, span id, etc.
ctx := tracer.InitServer("incoming_call", trace.Request{})
```

If you are starting from the client side:

```
import (
 "golang.org/x/net/context"
 "github.com/elodina/go-zipkin"
)

ctx := context.Background()

sendRPC(ctx)

func sendRPC(ctx context.Context) {

 // This adds "client send" annotation right away, and adds "client receive" annotation and collects formed span on defer
 defer fractionalTracer.TraceClient(&ctx, "rpc_call")(nil)

 if spanInfo, ok := zipkin.SampledSpanInfo(ctx); ok {
  // Here one may include a sampled span trace info into the call
 }
 
 // Here's where a call happens
}
```

To explicitly collect spans:

```
tracer.CollectFromContext(ctx)
```

## Examples

Simple example of usage is in the `example/main.go`
