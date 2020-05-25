package routes

import (
	"github.com/discordextremelist/api/ratelimit"
	"github.com/discordextremelist/api/util"
	"github.com/go-chi/chi"
	"net/http"
	"time"
)

var (
	botRatelimiter         ratelimit.Ratelimiter
	verifiedBotRatelimiter ratelimit.Ratelimiter
)

func Bot(w http.ResponseWriter, r *http.Request) {
	err, bot := util.LookupBot(chi.URLParam(r, "id"), false)
	if err != nil {
		util.WriteErrorResponse(w, err)
		return
	}
	util.WriteBotResponse(w, bot)
}

// TODO: Widget
func Widget(w http.ResponseWriter, _ *http.Request) {
	util.WriteNotImplementedResponse(w)
}

func InitBotRoutes() {
	botRatelimiter = ratelimit.NewRatelimiter(ratelimit.RatelimiterOptions{
		Limit:         10,
		Reset:         10000,
		RedisPrefix:   "rl_bots",
		TempBanAfter:  2,
		PermBanAfter:  2,
		TempBanLength: 24 * time.Hour,
	})
	verifiedBotRatelimiter = ratelimit.NewRatelimiter(ratelimit.RatelimiterOptions{
		Limit:         20,
		Reset:         10000,
		RedisPrefix:   "rl_verified_bots",
		TempBanAfter:  2,
		PermBanAfter:  2,
		TempBanLength: 24 * time.Hour,
	})
	util.Router.Route("/bot", func(r chi.Router) {
		r.Use(util.TokenValidator)
		r.Get("/{id}", Bot)
		r.Get("/{id}/widget", Widget)
		r.Post("/{id}/stats", Stats)
	})
}
