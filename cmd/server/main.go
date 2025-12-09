package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Divyamsirswal/flux-limiter/internal/limiter"
	"github.com/Divyamsirswal/flux-limiter/internal/middleware"
	"github.com/Divyamsirswal/flux-limiter/pkg/redis"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient, err := redis.NewClient(redisAddr)
	if err != nil {
		log.Fatalf("Critical: Redis connection failed: %v", err)
	}
	rateLimiter := limiter.NewRateLimiter(redisClient.Rdb)

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Pong! (Your request was allowed)\n"))
	})

	protectedHandler := middleware.RateLimitMiddleware(rateLimiter)(finalHandler)
	http.Handle("/ping", protectedHandler)
	log.Println("Flux Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
