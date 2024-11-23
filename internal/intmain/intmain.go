package intmain

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

func parseCLI() {
	verbose := false
	if len(os.Args) > 1 && os.Args[1] == "-v" {
		verbose = true
	}

	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.Debug("Verbose logging enabled")
}

func setupWebServer(port string) {
	http.HandleFunc("/{$}", rootHandler)
	http.HandleFunc("/app", appHandler)
	http.HandleFunc("/app/manifest.json", manifestHandler)
	http.HandleFunc("/app/icon.png", iconHandler)

	http.HandleFunc("/redirect", redirectHandler)
	http.HandleFunc("/getWebsiteInfos", getWebsiteInfoHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	logrus.Infof("Server listening on port %v...", port)

	err := http.ListenAndServe(":"+port, loggingMiddleware(http.DefaultServeMux))
	if err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}

func Main() {
	parseCLI()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	setupWebServer(port)
}
