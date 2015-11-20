// Copyright 2013 - by Jim Lawless
// License: MIT / X11
// See: http://www.mailsend-online.com/license2013.php

package main

import (
	"fmt"
	"github.com/elodina/go-zipkin/Godeps/_workspace/src/github.com/jimlawless/cfg"
	"log"
)

func main() {
	mymap := make(map[string]string)
	err := cfg.Load("test.cfg", mymap)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", mymap)
}
