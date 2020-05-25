package routes

import (
	"github.com/discordextremelist/api/ratelimit"
	"github.com/discordextremelist/api/util"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

func GetServer(w http.ResponseWriter, r *http.Request) {
	err, server := util.LookupServer(chi.URLParam(r, "id"), true)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			util.NotFound(w, r)
		} else {
			util.WriteErrorResponse(w, err)
		}
		return
	}
	util.WriteServerResponse(w, server)
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
