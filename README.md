# go-zipkin

Initial implementation of Zipkin tracing library for Go. Based on Spacemonkey Monitor: 
https://github.com/spacemonkeygo/monitor. Relies on using Google Context objects: http://blog.golang.org/context.

Simple example of usage is in the example/main.go

This library enables the following features:

 - Kafka span collector
 - Ability to collect spans to the collector whenever needed