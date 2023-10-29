package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "<h1>Hello Rejekts! <small>(version: %s)</small></h1>", version)
	})

	slog.Info("starting server", slog.String("version", version), slog.String("revision", revision), slog.String("date", revisionDate))

	err := http.ListenAndServe(":8080", router)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
