package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	limitergin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

const (
	rateLimitPeriod   = time.Minute
	rateLimitRequests = 60
)

// NewRateLimiter returns a Gin middleware that enforces a per-IP request cap
// on the URL shortener's public endpoints. For global limits across services,
// prefer configuring the gateway or reverse proxy layer instead.
func NewRateLimiter() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: rateLimitPeriod,
		Limit:  rateLimitRequests,
	}
	store := memory.NewStore()
	instance := limiter.New(store, rate)
	return limitergin.NewMiddleware(instance)
}
