package middleware



import (
	"log"
	"net/http"
	"time"
)



// LoggingMiddleware logs incoming HTTP requests
func Logging(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		// call next handler
		next.ServeHTTP(w, r)

		// log after response
		log.Printf(
			"%s %s %s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	})
}