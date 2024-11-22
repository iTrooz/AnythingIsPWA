package intmain

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// see https://gist.github.com/nstogner/2d6e122418ad3e21a175974e5c9bb36c
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := statusRecorder{w, 200}
		next.ServeHTTP(&rec, r)

		log.Infof("%v %v %v", rec.status, r.Method, r.URL.String())
	})
}
