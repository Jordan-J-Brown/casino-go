package httpx

import "net/http"

func IsClientError(code int) bool {
	return code >= http.StatusBadRequest && code < http.StatusInternalServerError
}

func IsServerError(code int) bool {
	return code >= http.StatusInternalServerError && code < 600
}
