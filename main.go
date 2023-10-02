package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, I'm Scribe!")
}

func main() {

	// Read in the config file.
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	port := fmt.Sprintf(":%s", viper.GetStringMapString("host")["port"])

	http.HandleFunc("/", handler)

	log.Printf("listening on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
