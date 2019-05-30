package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	olog "github.com/opentracing/opentracing-go/log"
)

var (
	s1 = rand.NewSource(time.Now().UnixNano())
	r1 = rand.New(s1)
)

func randSlice(n int) []byte {
	token := make([]byte, n)
	r1.Read(token)
	return token
}

// handleErr is a function that takes an error and an opentracing span and
// performs the necessary error handling functions such and printing relevant
// information and adding the error to the span.
func handleErr(err error, serverSpan opentracing.Span) {
	ext.Error.Set(serverSpan, true)
	serverSpan.LogFields(olog.Error(err))
	log.Println("ERROR:", err)
}
