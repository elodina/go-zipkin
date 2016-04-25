# Zipkin library for Go

This library focus is to provide maximum flexibility for creating [Zipkin](http://zipkin.io) traces inside the distributed application.
The library presumes the user is familiar with basic Zipkin concepts like annotations, spans, services, etc.
Zipkin spans are composed in a stateless fashion. It is up to you how to manage span entities inside the application. 
Kafka is used as a transport to transfer the completed spans to Zipkin collector.

## Quickstart
 
```go
// ...
import (
    "github.com/elodina/go-zipkin"
)

rate := 10 // tracing rate 1 of 10
brokerAddr := []string{"master:5000"} // Kafka broker endpoint

producer, err := zipkin.DefaultProducer(brokerAddr) if err != nil {
//...
}

tracer := zipkin.NewTracer("ServiceName", rate, producer, zipkin.LocalNetworkIP(), zipkin.DefaultPort(), zipkin.DefaultTopic())

//...
span := tracer.NewSpan("span_name")
span.ServerReceive()
// do work here
span.ServerSendAndCollect()
//...
```

## Examples

You may see the complete end-to-end example here: https://github.com/aShevc/go-zipkin-sample 
