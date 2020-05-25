package routes

import (
	"github.com/discordextremelist/api/ratelimit"
	"github.com/discordextremelist/api/util"
	"github.com/go-chi/chi"
	"net/http"
	"time"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	err, user := util.LookupUser(chi.URLParam(r, "id"), false)
	if err != nil {
		util.WriteErrorResponse(w, err)
		return
	}
	util.WriteUserResponse(w, user)
}

func InitUserRoutes() {
	// TODO: Decide on ratelimiting for users
	ratelimiter := ratelimit.NewRatelimiter(ratelimit.RatelimiterOptions{
		Limit:         10,
		Reset:         10000,
		RedisPrefix:   "rl_users",
		TempBanLength: 48 * time.Hour,
		TempBanAfter:  3,
		PermBanAfter:  2,
	})
	util.Router.Route("/user", func(r chi.Router) {
		r.Use(ratelimiter.Ratelimit)
		r.Get("/{id}", GetUser)
	})
}
