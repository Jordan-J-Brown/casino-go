//go:build profile

package main

import (
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

func init() {
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)
	go http.ListenAndServe(":6060", nil)
}
