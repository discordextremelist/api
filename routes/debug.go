package routes

import (
	"context"
	"github.com/discordextremelist/api/entities"
	"github.com/discordextremelist/api/util"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"os"
)

func Debug(w http.ResponseWriter, r *http.Request) {
	if !util.Dev {
		token := r.URL.Query().Get("token")
		if token == "" {
			entities.WriteJson(403, w, map[string]interface{}{})
			return
		}
		err := util.Database.Mongo.Collection("adminTokens").FindOne(context.TODO(), bson.M{"token": token}).Err()
		if err != nil {
			entities.WriteJson(403, w, map[string]interface{}{})
			return
		}
	}
	debug(w)
}

func debug(w http.ResponseWriter) {
	hostname, _ := os.Hostname()
	entities.WritePrettyJson(200, w, &entities.DebugStatistics{
		RedisPing: util.Database.PingRedis(),
		MongoPing: util.Database.PingMongo(),
		Node:      util.Node,
		LookupTimes: entities.LookupTimes{
			Mongo: entities.MongoLookupTimes,
			Redis: entities.RedisLookupTimes,
		},
		ResponseTimes: entities.ResponseTimes,
		Hostname:      hostname,
	})
}

func InitDebugRoutes() {
	util.Router.Get("/debug", Debug)
}
