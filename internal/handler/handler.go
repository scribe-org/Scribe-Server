package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, I'm Scribe!")
}

// HandleRequests handles incoming HTTP requests.
func HandleRequests() {
	host := viper.GetStringMapString("host")
	port := fmt.Sprintf(":%s", host["port"])

	// Setup root handler.
	http.HandleFunc("/", hello)

	// Setup /files handler.
	fileSystem := viper.GetString("fileSystem")
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(fileSystem))))

	// Start serving requests.
	log.Printf("listening on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
