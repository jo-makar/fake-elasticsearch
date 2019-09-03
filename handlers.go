package main

import (
	"net/http"
)

type RootHandler struct {
	*State
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// FIXME STOPPED
	w.WriteHeader(http.StatusOK)
}
