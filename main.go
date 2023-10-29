package main

import (
	"fmt"
	"net/http"
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "<h1>Hello Rejekts! <small>(version: %s)</small></h1>", version)
	})

	err := http.ListenAndServe(":8080", router)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
