package routes

import (
	"github.com/discordextremelist/api/entities"
	"github.com/discordextremelist/api/ratelimit"
	"github.com/discordextremelist/api/util"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	err, user := entities.LookupUser(chi.URLParam(r, "id"), true)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			entities.NotFound(w, r)
		} else {
			sentry.CaptureException(err)
			entities.WriteErrorResponse(w)
		}
		return
	}
	entities.WriteUserResponse(w, user)
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
