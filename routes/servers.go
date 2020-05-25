package routes

import (
	"github.com/discordextremelist/api/ratelimit"
	"github.com/discordextremelist/api/util"
	"github.com/go-chi/chi"
	"net/http"
	"time"
)

func GetServer(w http.ResponseWriter, _ *http.Request) {
	util.WriteNotImplementedResponse(w)
}

func InitServerRoutes() {
	// TODO: Decide on ratelimiting for servers
	ratelimiter := ratelimit.NewRatelimiter(ratelimit.RatelimiterOptions{
		Limit:         10,
		Reset:         10000,
		RedisPrefix:   "rl_servers",
		TempBanLength: 48 * time.Hour,
		TempBanAfter:  3,
		PermBanAfter:  2,
	})
	util.Router.Route("/server", func(r chi.Router) {
		r.Use(ratelimiter.Ratelimit)
		r.Get("/{id}", GetServer)
	})
}
