package ratelimit

import (
	"context"
	"github.com/discordextremelist/api/entities"
	"github.com/discordextremelist/api/util"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var TempBanReset = 7 * 24 * time.Hour

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
	Cache         map[string]*Ratelimit
	Limit         int
	Reset         int
	Mutex         *sync.Mutex
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

func NewRatelimiter(opts RatelimiterOptions) Ratelimiter {
	rl := Ratelimiter{
		Cache:         make(map[string]*Ratelimit),
		Limit:         opts.Limit,
		Reset:         opts.Reset,
		Mutex:         &sync.Mutex{},
		NextReset:     time.Now().Add(time.Duration(opts.Reset) * time.Millisecond),
		RPrefix:       opts.RedisPrefix,
		TempBanLength: opts.TempBanLength,
		TempBanAfter:  opts.TempBanAfter,
		PermBanAfter:  opts.PermBanAfter,
	}
	s := time.Now()
	count := rl.cacheAll()
	log.WithField("ratelimiter", opts.RedisPrefix).Debugf("Caching %d ratelimit(s) took: %v", count, time.Now().Sub(s))
	go rl.resetRatelimits()
	go rl.resetTempBans()
	return rl
}

func (r *Ratelimiter) cacheAll() int {
	results, err := util.Database.Redis.HGetAll(context.TODO(), r.RPrefix).Result()
	if err != nil {
		log.WithField("ratelimiter", r.RPrefix).Fatalf("Failed to get ratelimits!")
	}
	for key, val := range results {
		ratelimit := &Ratelimit{}
		_ = util.Json.UnmarshalFromString(val, ratelimit)
		r.overwrite(key, *ratelimit)
	}
	return len(results)
}

func (r *Ratelimiter) resetRatelimits() {
	for {
		select {
		case <-time.After(time.Duration(r.Reset) * time.Millisecond):
			{
				r.NextReset = time.Now().Add(time.Duration(r.Reset) * time.Millisecond)
				for k := range r.Cache {
					r.reset(k)
				}
			}
		}
	}
}

func (r *Ratelimiter) cacheRatelimit(key string, ratelimit Ratelimit) {
	if ratelimit.PermBannedAt > 0 {
		return
	}
	if util.Database.IsRedisOpen() {
		str, _ := util.Json.MarshalToString(&ratelimit)
		util.Database.Redis.HMSet(context.TODO(), r.RPrefix, key, str)
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
				for k, v := range r.Cache {
					if v.PermBannedAt > 0 {
						return
					}
					v.Unpatch()
					v.AfterClearCount = 0
					r.overwrite(k, *v)
				}
			}
		}
	}
}

func (r *Ratelimiter) getRatelimit(key string) *Ratelimit {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	rl := r.Cache[key]
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
	r.Cache[key] = rl
	r.cacheRatelimit(key, *rl)
	return rl
}

func (r *Ratelimiter) reset(key string) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	rl := r.Cache[key]
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
	r.Cache[key] = rl
	r.cacheRatelimit(key, *rl)
}

func (r *Ratelimiter) overwrite(key string, ratelimit Ratelimit) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.Cache[key] = &ratelimit
}

func (r *Ratelimiter) Ratelimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if util.CheckIP(req.RemoteAddr) {
			next.ServeHTTP(writer, req)
			return
		}
		ratelimit := r.getRatelimit(req.RemoteAddr)
		headers := writer.Header()
		if ratelimit.TotalBans > 0 && (ratelimit.TempBannedAt > 0 || ratelimit.PermBannedAt > 0) {
			headers.Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusForbidden)
			if !ratelimit.TempBan {
				util.Json.NewEncoder(writer).Encode(entities.PermBannedError)
			} else {
				util.Json.NewEncoder(writer).Encode(entities.TempBannedError)
			}
			return
		}
		left := r.Limit - ratelimit.Current
		if left <= 0 {
			headers.Set("Content-Type", "application/json")
			headers.Set("Retry-After", strconv.FormatInt(r.NextReset.Sub(time.Now()).Milliseconds(), 10))
			writer.WriteHeader(http.StatusTooManyRequests)
			util.Json.NewEncoder(writer).Encode(entities.RatelimitedError)
			return
		}
		headers.Set("X-RateLimit-Limit", strconv.Itoa(r.Limit))
		headers.Set("X-RateLimit-Remaining", strconv.Itoa(left))
		headers.Set("X-RateLimit-Reset", strconv.FormatInt(r.NextReset.Unix()*1000, 10))
		headers.Set("X-RateLimit-Bucket", strings.Replace(r.RPrefix, "rl_", "", 1))
		next.ServeHTTP(writer, req)
	})
}
