package middlewares

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/AdhityaRamadhanus/userland/common/contextkey"
	redis "github.com/go-redis/redis"
	ratelimiter "github.com/teambition/ratelimiter-go"
)

type redisRateLimitClient struct {
	*redis.Client
}

func (c *redisRateLimitClient) RateDel(key string) error {
	return c.Del(key).Err()
}
func (c *redisRateLimitClient) RateEvalSha(sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return c.EvalSha(sha1, keys, args...).Result()
}
func (c *redisRateLimitClient) RateScriptLoad(script string) (string, error) {
	return c.ScriptLoad(script).Result()
}

type RateLimiter struct {
	redisRateLimitClient *redisRateLimitClient
	limiter              *ratelimiter.Limiter
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redisRateLimitClient: &redisRateLimitClient{redisClient},
	}
}

func (r *RateLimiter) Limit(maxFreq int, duration time.Duration) func(nextHandler http.HandlerFunc) http.HandlerFunc {
	r.limiter = ratelimiter.New(ratelimiter.Options{
		Max:      maxFreq,
		Duration: duration, // limit to 1000 requests in 1 minute.
		Client:   r.redisRateLimitClient,
	})
	return func(nextHandler http.HandlerFunc) http.HandlerFunc {
		return (http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			clientInfo := req.Context().Value(contextkey.ClientInfo).(map[string]interface{})
			ctx, err := r.limiter.Get(fmt.Sprintf("%s:%s", req.URL.Path, clientInfo["ip"].(string)))
			if err != nil {
				http.Error(res, err.Error(), 500)
				return
			}

			header := res.Header()
			header.Set("X-Ratelimit-Limit", strconv.FormatInt(int64(ctx.Total), 10))
			header.Set("X-Ratelimit-Remaining", strconv.FormatInt(int64(ctx.Remaining), 10))
			header.Set("X-Ratelimit-Reset", strconv.FormatInt(ctx.Reset.Unix(), 10))

			if ctx.Remaining < 0 {
				after := int64(ctx.Reset.Sub(time.Now())) / 1e9
				header.Set("Retry-After", strconv.FormatInt(after, 10))
				res.WriteHeader(429)
				fmt.Fprintf(res, "Rate limit exceeded, retry in %d seconds.\n", after)
				return
			}

			nextHandler(res, req)
		}))
	}
}
