package main

import (
	"net/http"
	"strings"
)

func GetIp(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	} else {
		return strings.Split(r.RemoteAddr, ":")[0]
	}
}
