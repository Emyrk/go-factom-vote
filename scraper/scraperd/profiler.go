package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
)

// StartProfiler runs the go pprof tool
// `go tool pprof http://localhost:6060/debug/pprof/profile`
// https://golang.org/pkg/net/http/pprof/
func StartProfiler(expose bool) {
	_ = log.Print
	//runtime.MemProfileRate = mpr
	pre := "localhost"
	if expose {
		pre = ""
	}
	log.Println(http.ListenAndServe(fmt.Sprintf("%s:%d", pre, 6060), nil))
	//runtime.SetBlockProfileRate(100000)
}
