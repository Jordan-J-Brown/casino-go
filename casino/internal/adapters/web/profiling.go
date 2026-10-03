//go:build local

//This will only run when using the build tags "-tags local"

package web

import (
	"net/http"
	_ "net/http/pprof"
)

func init() {
	go http.ListenAndServe(":6060", nil)
}
