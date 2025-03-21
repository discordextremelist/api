package routes

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/discordextremelist/api/entities"
	"github.com/discordextremelist/api/ratelimit"
	"github.com/discordextremelist/api/util"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	botsRatelimiter       *ratelimit.Ratelimiter
	premiumBotRatelimiter *ratelimit.Ratelimiter
)

func Bot(w http.ResponseWriter, r *http.Request) {
	err, bot := entities.LookupBot(chi.URLParam(r, "id"), true)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			entities.NotFound(w, r)
		} else {
			sentry.CaptureException(err)
			entities.WriteErrorResponse(w)
		}
		return
	}
	entities.WriteBotResponse(w, bot)
}

func Bots(w http.ResponseWriter, _ *http.Request) {
	err, bots := entities.GetAllBots(true)
	if err != nil {
		sentry.CaptureException(err)
		entities.WriteErrorResponse(w)
		return
	}
	entities.WriteJson(200, w, entities.APIResponseBots{
		Error:  false,
		Status: 200,
		Bots:   bots,
	})
}

// TODO: Widget
func Widget(w http.ResponseWriter, _ *http.Request) {
	entities.WriteNotImplementedResponse(w)
}

type StatsRequest struct {
	GuildCount int `json:"guildCount"`
	ShardCount int `json:"shardCount"`
}

func UpdateStats(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get(util.ContentType), "application/json") {
		entities.WriteJson(400, w, entities.BadContentType)
	} else {
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			sentry.CaptureException(err)
			entities.WriteErrorResponse(w)
			return
		}
		var body StatsRequest
		err = json.Unmarshal(bytes, &body)
		if err != nil {
			sentry.CaptureException(err)
			entities.WriteErrorResponse(w)
			return
		}
		err, bot := entities.LookupBot(chi.URLParam(r, "id"), false)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				entities.NotFound(w, r)
			} else {
				sentry.CaptureException(err)
				entities.WriteErrorResponse(w)
			}
			return
		}
		if !util.Dev && (r.Header.Get(util.Authorization) != bot.Token) {
			entities.BadAuth(w, r)
			return
		}
		set := bson.M{}
		if body.GuildCount > 0 {
			bot.ServerCount = body.GuildCount
			set["serverCount"] = body.GuildCount
		} else {
			set["serverCount"] = bot.ServerCount
		}
		if body.ShardCount > 0 {
			bot.ShardCount = body.ShardCount
			set["shardCount"] = body.ShardCount
		} else {
			set["shardCount"] = bot.ShardCount
		}
		marshaled, err := json.Marshal(bot)
		if err != nil {
			sentry.CaptureException(err)
			entities.WriteErrorResponse(w)
			return
		}
		err = util.Database.Redis.HMSet(context.TODO(), "bots", bot.ID, string(marshaled)).Err()
		if err != nil {
			sentry.CaptureException(err)
			entities.WriteErrorResponse(w)
			return
		}
		_, err = util.Database.Mongo.Collection("bots").UpdateOne(context.TODO(), bson.M{"_id": bot.ID}, bson.D{{"$set", set}})
		if err != nil {
			sentry.CaptureException(err)
			entities.WriteErrorResponse(w)
			return
		}
		entities.WriteJson(200, w, map[string]interface{}{"status": 200, "error": false, "updated": body})
	}
}

func InitBotRoutes() {
	botsRatelimiter = ratelimit.NewRatelimiter(ratelimit.RatelimiterOptions{
		Limit:         10,
		Reset:         60000,
		RedisPrefix:   "rl_bots",
		TempBanAfter:  3,
		PermBanAfter:  3,
		TempBanLength: 24 * time.Hour,
	})
	premiumBotRatelimiter = ratelimit.NewRatelimiter(ratelimit.RatelimiterOptions{
		Limit:         20,
		Reset:         10000,
		RedisPrefix:   "rl_premium_bots",
		TempBanAfter:  4,
		PermBanAfter:  4,
		TempBanLength: 24 * time.Hour,
	})
	util.Router.Route("/bots", func(r chi.Router) {
		r.Use(botsRatelimiter.Ratelimit)
		r.Get("/", Bots)
	})
	util.Router.Route("/bot/{id}", func(r chi.Router) {
		r.Use(entities.TokenValidator)
		r.Use(func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				err, bot := entities.LookupBot(chi.URLParam(request, "id"), true)
				if err != nil {
					botsRatelimiter.Ratelimit(handler).ServeHTTP(writer, request)
				} else {
					if bot.Status.Premium {
						premiumBotRatelimiter.Ratelimit(handler).ServeHTTP(writer, request)
					} else {
						botsRatelimiter.Ratelimit(handler).ServeHTTP(writer, request)
					}
				}
			})
		})
		r.Get("/", Bot)
		r.Get("/widget", Widget)
		r.Post("/stats", UpdateStats)
	})
}
