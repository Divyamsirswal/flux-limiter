package middleware

import (
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Divyamsirswal/flux-limiter/internal/limiter"
)

func RateLimitMiddleware(l *limiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = strings.Split(r.RemoteAddr, ":")[0]
			}

			limit := 5
			window := 10 * time.Second

			allowed, err := l.Allow(r.Context(), ip, limit, window)
			if err != nil {
				log.Printf("Rate limit error: %v", err)
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				w.WriteHeader(http.StatusTooManyRequests) 
				w.Write([]byte("429: Too Many Requests. Slow down, buddy.\n"))
				return 
			}

			next.ServeHTTP(w, r)
		})
	}
}
