package intmain

import (
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/sirupsen/logrus"
)

func handleVerboseLogging(config *Config) {
	if config.Verbose || (len(os.Args) > 1 && os.Args[1] == "-v") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Debug("Verbose logging enabled")
}

func setupWebServer(config *Config) {
	http.HandleFunc("/{$}", rootHandler)
	http.HandleFunc("/app", appHandler)
	http.HandleFunc("/app/manifest.json", manifestHandler)
	http.HandleFunc("/app/icon.png", iconHandler)

	http.HandleFunc("/redirect", redirectHandler)
	http.HandleFunc("/getWebsiteInfos", withRateLimit(getWebsiteInfoHandler, config.RateLimit, config.RateLimitInterval))

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	logrus.Infof("Server listening on port %v...", config.Port)

	err := http.ListenAndServe(":"+config.Port, loggingMiddleware(http.DefaultServeMux))
	if err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}

type Config struct {
	Verbose           bool          `env:"VERBOSE" envDefault:"false"`
	Port              string        `env:"PORT" envDefault:"8080"`
	RateLimit         int           `env:"RATE_LIMIT" envDefault:"5"`
	RateLimitInterval time.Duration `env:"RATE_LIMIT_INTERVAL" envDefault:"5s"`
}

func Main() {
	config := &Config{}
	if err := env.Parse(config); err != nil {
		logrus.Fatalf("Failed to parse environment variables: %v", err)
	}
	logrus.Infof("Environment config: %+v\n", *config)

	handleVerboseLogging(config)

	setupWebServer(config)
}
