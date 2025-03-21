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

func GetTemplate(w http.ResponseWriter, r *http.Request) {
	err, template := entities.LookupTemplate(chi.URLParam(r, "id"))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			entities.NotFound(w, r)
		} else {
			sentry.CaptureException(err)
			entities.WriteErrorResponse(w)
		}
		return
	}
	entities.WriteTemplateResponse(w, template)
}

func InitTemplateRoutes() {
	ratelimiter := ratelimit.NewRatelimiter(ratelimit.RatelimiterOptions{
		Limit:         10,
		Reset:         10000,
		RedisPrefix:   "rl_templates",
		TempBanLength: 48 * time.Hour,
		TempBanAfter:  3,
		PermBanAfter:  2,
	})
	util.Router.Route("/template", func(r chi.Router) {
		r.Use(ratelimiter.Ratelimit)
		r.Get("/{id}", GetTemplate)
	})
}
