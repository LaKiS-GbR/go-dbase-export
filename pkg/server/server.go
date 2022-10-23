package server

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Start the server
func Start(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", IndexHandler)
	mux.HandleFunc("/export", ExportHandler)

	server := http.Server{
		Addr:              fmt.Sprintf(":%v", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start the server
	log.Fatal(server.ListenAndServe())
}
