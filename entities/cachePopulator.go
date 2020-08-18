package entities

import (
	"context"
	"github.com/discordextremelist/api/util"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

var collections = []string{"bots", "users", "servers", "templates"}

func PopulateDevCache() {
	logrus.Info("Populating redis cache for development use...")
	for _, col := range collections {
		_ = util.Database.Redis.Del(context.TODO(), col)
		start := time.Now()
		logrus.Infof("Populating redis cache for collection %s...", col)
		var mongoEntities []bson.M
		cursor, err := util.Database.Mongo.Collection(col).Find(context.TODO(), bson.M{})
		if err != nil {
			logrus.Errorf("Failed querying database for collection %s: %s", col, err.Error())
			continue
		}
		for cursor.Next(context.TODO()) {
			var entity bson.M
			err = cursor.Decode(&entity)
			entity["id"] = entity["_id"].(string)
			mongoEntities = append(mongoEntities, entity)
		}
		_ = cursor.Close(context.TODO())
		var toSet []string
		for _, entity := range mongoEntities {
			id := entity["_id"].(string)
			delete(entity, "_id")
			marshaled, err := util.Json.MarshalToString(&entity)
			if err != nil {
				logrus.Errorf("Failed to marshal an entity for cache %s (col: %s, err: %s)", id, col, err.Error())
				continue
			}
			toSet = append(toSet, id)
			toSet = append(toSet, marshaled)
		}
		err = util.Database.Redis.HMSet(context.TODO(), col, toSet).Err()
		if err != nil {
			logrus.Errorf("Failed to populate cache for redis map %s", col, err.Error())
			continue
		}
		logrus.Infof("Took %s to populate cache!", time.Now().Sub(start))
	}
}
