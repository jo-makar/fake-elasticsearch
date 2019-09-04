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

	state := NewState()

	http.Handle("/", &RootHandler{state})

	// Wrap the default handler/multiplexer to log incoming requests
	mux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			log.Printf("%s %s %s\n", r.Method, r.RequestURI, GetIp(r))
		}()
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	server := http.Server{
		   Addr: fmt.Sprintf("%s:%d", addr, port),
		Handler: mux,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s := <-sigChan
		log.Printf("terminating signal received (%d)\n", s)

		// Cause ListenAndServe to return with ErrServerClosed
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			log.Panic(err)
		}
	}
}
