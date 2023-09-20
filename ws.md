package main

import (
	"github.com/pretty66/websocketproxy"
	"log"
	"net/http"
)

func main() {
	wp, err := websocketproxy.NewProxy("ws://107.150.5.98:13839/ws", func(r *http.Request) error {
		// Permission to verify
		// r.Header.Set("Cookie", "----")
		// Source of disguise
		// r.Header.Set("Origin", "http://82.157.123.54:9010")
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	// proxy path
	http.HandleFunc("/ws", wp.Proxy)
	http.ListenAndServe(":9696", nil)
}
