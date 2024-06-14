package intmain

import (
	"fmt"
	"net/http"
	"os"
)

func Main() {
	http.HandleFunc("/{$}", rootHandler)
	http.HandleFunc("/app", appHandler)
	http.HandleFunc("/app/manifest.json", manifestHandler)
	http.HandleFunc("/app/icon.png", iconHandler)

	http.HandleFunc("/redirect", redirectHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server listening on port %s...\n", port)
	http.ListenAndServe(":"+port, loggingMiddleware(http.DefaultServeMux))
}
