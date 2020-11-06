package routes

import (
	"context"
	"github.com/discordextremelist/api/entities"
	"github.com/discordextremelist/api/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net/http"
	"os"
	"time"
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
	redisPing := time.Now()
	var redisPingEnd int64
	err := util.Database.Redis.Ping(context.TODO()).Err()
	if err != nil {
		redisPingEnd = -1
	} else {
		redisPingEnd = time.Since(redisPing).Milliseconds()
	}
	mongoPingStart := time.Now()
	var mongoPingEnd int64
	err = util.Database.Mongo.Client().Ping(context.TODO(), readpref.Primary())
	if err != nil {
		mongoPingEnd = -1
	} else {
		mongoPingEnd = time.Since(mongoPingStart).Milliseconds()
	}
	hostname, _ := os.Hostname()
	entities.WritePrettyJson(200, w, &entities.DebugStatistics{
		RedisPing: redisPingEnd,
		MongoPing: mongoPingEnd,
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
