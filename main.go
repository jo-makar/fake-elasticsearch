package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var port int
	var addr string

	flag.IntVar(&port, "port", 9200, "port number")
	flag.StringVar(&addr, "addr", "", "bound address")
	flag.Parse()

	log.Printf("using %s:%d\n", addr, port)

	state, err := NewState()
	if err != nil {
		log.Panic(err)
	}

	http.Handle("/", &RootHandler{state})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	server := http.Server{ Addr: fmt.Sprintf("%s:%d", addr, port) }

	go func() {
		s := <-sigChan
		log.Printf("terminating signal received (%d)\n", s)

		// This causes ListenAndServe to return with ErrServerClosed
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			log.Panic(err)
		}
	}
}
