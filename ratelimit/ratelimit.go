package ratelimit

import (
	"context"
	"encoding/json"
	"github.com/discordextremelist/api/entities"
	"github.com/discordextremelist/api/util"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	DefaultRatelimit = &Ratelimit{
		Current:         0,
		AfterClearCount: 0,
		TempBan:         false,
		TempBannedAt:    0,
		PermBannedAt:    0,
		TotalBans:       0,
	}
	TempBanReset = 7 * 24 * time.Hour
)

type Ratelimit struct {
	Current         int   `json:"current"`
	AfterClearCount int   `json:"after_clear_count"`
	TempBan         bool  `json:"temp_ban"`
	TempBannedAt    int64 `json:"temp_banned_at"`
	PermBannedAt    int64 `json:"perm_banned_at"`
	TotalBans       int   `json:"total_bans"`
}

func (r *Ratelimit) PatchTemp() {
	if !r.TempBan {
		r.TempBan = true
	}
	r.TempBannedAt = time.Now().UnixNano()
	r.TotalBans++
}

func (r *Ratelimit) Unpatch() {
	r.TempBan = false
	r.TempBannedAt = 0
}

func (r *Ratelimit) PatchPerm() {
	r.TempBan = false
	r.TempBannedAt = 0
	r.PermBannedAt = time.Now().UnixNano()
}

type Ratelimiter struct {
	Limit         int
	Reset         int
	NextReset     time.Time
	RPrefix       string
	TempBanLength time.Duration
	TempBanAfter  int
	PermBanAfter  int
}

type RatelimiterOptions struct {
	Limit         int
	Reset         int
	RedisPrefix   string
	TempBanLength time.Duration
	TempBanAfter  int
	PermBanAfter  int
}

func NewRatelimiter(opts RatelimiterOptions) *Ratelimiter {
	rl := &Ratelimiter{
		Limit:         opts.Limit,
		Reset:         opts.Reset,
		NextReset:     time.Now().Add(time.Duration(opts.Reset) * time.Millisecond),
		RPrefix:       opts.RedisPrefix,
		TempBanLength: opts.TempBanLength,
		TempBanAfter:  opts.TempBanAfter,
		PermBanAfter:  opts.PermBanAfter,
	}
	s := time.Now()
	count := util.Database.Redis.HLen(context.TODO(), opts.RedisPrefix).Val()
	log.WithField("ratelimiter", opts.RedisPrefix).Debugf("Took %s to get %d ratelimits!", time.Now().Sub(s), count)
	go rl.resetRatelimits()
	go rl.resetTempBans()
	return rl
}

func (r *Ratelimiter) getAll() map[string]*Ratelimit {
	ratelimits := make(map[string]*Ratelimit)
	results, err := util.Database.Redis.HGetAll(context.TODO(), r.RPrefix).Result()
	if err != nil {
		log.WithField("ratelimiter", r.RPrefix).Fatalf("Failed to get ratelimits!")
		return ratelimits
	}
	for k, val := range results {
		ratelimit := &Ratelimit{}
		_ = json.Unmarshal([]byte(val), ratelimit)
		ratelimits[k] = ratelimit
	}
	return ratelimits
}

func (r *Ratelimiter) resetRatelimits() {
	for {
		select {
		case <-time.After(time.Duration(r.Reset) * time.Millisecond):
			{
				r.NextReset = time.Now().Add(time.Duration(r.Reset))
				for k := range r.getAll() {
					r.reset(k)
				}
			}
		}
	}
}

func (r *Ratelimiter) cacheRatelimit(key string, ratelimit *Ratelimit) {
	if ratelimit.PermBannedAt > 0 {
		return
	}
	if util.Database.IsRedisOpen() {
		byteArr, _ := json.Marshal(&ratelimit)
		util.Database.Redis.HMSet(context.TODO(), r.RPrefix, key, string(byteArr))
	}
}

func (r *Ratelimiter) HasExpired(ratelimit *Ratelimit) bool {
	return (time.Now().UnixNano() - ratelimit.TempBannedAt) >= r.TempBanLength.Nanoseconds()
}

func (r *Ratelimiter) resetTempBans() {
	for {
		select {
		case <-time.After(TempBanReset):
			{
				for k, v := range r.getAll() {
					if v.PermBannedAt > 0 {
						return
					}
					v.Unpatch()
					v.AfterClearCount = 0
					r.cacheRatelimit(k, v)
				}
			}
		}
	}
}

func (r *Ratelimiter) getRatelimit(key string) *Ratelimit {
	res, err := util.Database.Redis.HGet(context.TODO(), r.RPrefix, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.cacheRatelimit(key, DefaultRatelimit)
			return DefaultRatelimit
		}
		return DefaultRatelimit
	}
	var rl *Ratelimit
	err = json.Unmarshal([]byte(res), &rl)
	if err != nil {
		return DefaultRatelimit
	}
	if rl == nil {
		rl = &Ratelimit{
			Current:         0,
			AfterClearCount: 0,
			TempBan:         false,
			TempBannedAt:    0,
			PermBannedAt:    0,
			TotalBans:       0,
		}
	}
	rl.Current++
	r.cacheRatelimit(key, rl)
	return rl
}

func (r *Ratelimiter) reset(key string) {
	rl := r.getRatelimit(key)
	if rl.PermBannedAt < 1 && rl.TotalBans == r.PermBanAfter {
		rl.PatchPerm()
	}
	if rl.TotalBans > 0 && (rl.TempBannedAt > 0 || rl.PermBannedAt > 0) {
		if rl.TempBan && r.HasExpired(rl) {
			rl.Unpatch()
			rl.AfterClearCount = 0
		}
	}
	if (r.Limit-rl.Current) <= 0 && !rl.TempBan {
		if rl.PermBannedAt > 0 {
			return
		}
		rl.AfterClearCount++
		if rl.AfterClearCount >= r.TempBanAfter {
			rl.PatchTemp()
		}
	}
	rl.Current = 0
	r.cacheRatelimit(key, rl)
}

func (r *Ratelimiter) Ratelimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		ratelimit := r.getRatelimit(req.RemoteAddr)
		headers := writer.Header()
		if ratelimit.TotalBans > 0 && (ratelimit.TempBannedAt > 0 || ratelimit.PermBannedAt > 0) {
			headers.Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusForbidden)
			if !ratelimit.TempBan {
				json.NewEncoder(writer).Encode(entities.PermBannedError)
			} else {
				json.NewEncoder(writer).Encode(entities.TempBannedError)
			}
			return
		}
		left := r.Limit - ratelimit.Current
		if left <= 0 {
			headers.Set("Content-Type", "application/json")
			headers.Set("Retry-After", strconv.FormatInt(time.Now().Sub(r.NextReset).Milliseconds(), 10))
			writer.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(writer).Encode(entities.RatelimitedError)
			return
		}
		headers.Set("X-RateLimit-Limit", strconv.Itoa(r.Limit))
		headers.Set("X-RateLimit-Remaining", strconv.Itoa(left))
		headers.Set("X-RateLimit-Reset", strconv.FormatInt(r.NextReset.Unix()*1000, 10))
		headers.Set("X-RateLimit-Bucket", strings.Replace(r.RPrefix, "rl_", "", 1))
		next.ServeHTTP(writer, req)
	})
}
