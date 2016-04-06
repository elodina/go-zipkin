# Zipkin library for Go

Initial implementation of Zipkin tracing library for Go. Specifically created for collecting Zipkin spans using Kafka as a transport. 

This library enables the following features:

 - Kafka span collector
 - Ability to collect spans to the collector whenever needed
 
## Quickstart
 
```go
// ...
import (
    "github.com/elodina/go-zipkin"
    "github.com/elodina/siesta-producer"
)

rate := 10 // tracing rate 1 of 10
kafkaProducer := producer.NewKafkaProducer(...)

tracer := zipkin.NewTracer("ServiceName", "host", rate, kafkaProducer)
//...
span := tracer.NewSpan("span_name")
span.ServerReceive()
// do work here
span.ServerSendAndCollect()
//...
```

