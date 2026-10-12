//go:build profile

//This will only run when using the build tags "-tags profile"

package web

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
