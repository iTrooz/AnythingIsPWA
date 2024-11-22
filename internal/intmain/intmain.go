package intmain

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

func Main() {
	http.HandleFunc("/{$}", rootHandler)
	http.HandleFunc("/app", appHandler)
	http.HandleFunc("/app/manifest.json", manifestHandler)
	http.HandleFunc("/app/icon.png", iconHandler)

	http.HandleFunc("/redirect", redirectHandler)
	http.HandleFunc("/getWebsiteInfos", getWebsiteInfoHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logrus.Infof("Server listening on port %s...", port)

	err := http.ListenAndServe(":"+port, loggingMiddleware(http.DefaultServeMux))
	if err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
