package util

import (
	"context"
	"github.com/discordextremelist/api/database"
	"github.com/discordextremelist/api/entities"
	"github.com/go-chi/chi"
	"github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	Database = database.NewManager()
	Json     = jsoniter.ConfigFastest
	Router   chi.Router
	Dev      bool
	// TODO: If we ever decide to add user tokens, we can get the current users rank with ease, potentially modifying the data of entities.CleanupBot
	fakeRank = entities.UserRank{
		Admin:      false,
		Assistant:  false,
		Mod:        false,
		Premium:    false,
		Tester:     false,
		Translator: false,
		Covid:      false,
	}
)

func mongoLookupBot(id string) (error, *entities.Bot) {
	res := Database.Mongo.Collection("bots").FindOne(context.TODO(), bson.M{"_id": id})
	if res.Err() != nil {
		return res.Err(), nil
	}
	bot := entities.Bot{}
	if err := res.Decode(&bot); err != nil {
		return err, nil
	}
	return nil, entities.CleanupBot(fakeRank, &bot)
}

func LookupBot(id string, clean bool) (error, *entities.Bot) {
	redisBot, err := Database.Redis.HGet("bots", id).Result()
	if err != nil {
		if redisBot == "" {
			err, bot := mongoLookupBot(id)
			if err != nil {
				log.Errorf("Fallback for MongoDB failed for LookupBot(%s): %v", id, err.Error())
				return LookupError, nil
			} else {
				if clean {
					bot = entities.CleanupBot(fakeRank, bot)
				}
				return nil, bot
			}
		}
		bot := &entities.Bot{}
		err = Json.UnmarshalFromString(redisBot, &bot)
		if err != nil {
			log.Errorf("Json parsing failed for LookupBot(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			if clean {
				bot = entities.CleanupBot(fakeRank, bot)
			}
			return nil, bot
		}
	} else {
		err, bot := mongoLookupBot(id)
		if err != nil {
			log.Errorf("Fallback for MongoDB failed for LookupBot(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			if clean {
				bot = entities.CleanupBot(fakeRank, bot)
			}
			return nil, bot
		}
	}
}

func mongoLookupUser(id string) (error, *entities.User) {
	res := Database.Mongo.Collection("users").FindOne(context.TODO(), bson.M{"_id": id})
	if res.Err() != nil {
		return res.Err(), nil
	}
	user := entities.User{}
	if err := res.Decode(&user); err != nil {
		return err, nil
	}
	return nil, entities.CleanupUser(fakeRank, &user)
}

func LookupUser(id string, clean bool) (error, *entities.User) {
	redisUser, err := Database.Redis.HGet("users", id).Result()
	if err != nil {
		if redisUser == "" {
			err, user := mongoLookupUser(id)
			if err != nil {
				log.Errorf("Fallback for MongoDB failed for LookupUser(%s): %v", id, err.Error())
				return LookupError, nil
			} else {
				if clean {
					user = entities.CleanupUser(fakeRank, user)
				}
				return nil, user
			}
		}
		user := &entities.User{}
		err = Json.UnmarshalFromString(redisUser, &user)
		if err != nil {
			log.Errorf("Json parsing failed for LookupBot(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			return nil, entities.CleanupUser(fakeRank, user)
		}
	} else {
		err, user := mongoLookupUser(id)
		if err != nil {
			log.Errorf("Fallback for MongoDB failed for LookupUser(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			if clean {
				user = entities.CleanupUser(fakeRank, user)
			}
			return nil, user
		}
	}
}

func GetAllBots(clean bool) (error, []entities.Bot) {
	redisBots, err := Database.Redis.HVals("bots").Result()
	if err != nil && len(redisBots) > 0 {
		actual := make([]entities.Bot, len(redisBots))
		for _, str := range redisBots {
			bot := entities.Bot{}
			err = Json.UnmarshalFromString(str, &bot)
			if err != nil {
				continue
			}
			if clean {
				bot = *entities.CleanupBot(fakeRank, &bot)
			}
			actual = append(actual, bot)
		}
		return nil, actual
	} else {
		cursor, err := Database.Mongo.Collection("bots").Find(context.TODO(), bson.M{})
		if err != nil {
			return err, nil
		}
		var actual []entities.Bot
		defer cursor.Close(context.TODO())
		for cursor.Next(context.TODO()) {
			bot := entities.Bot{}
			err = cursor.Decode(&bot)
			if err != nil {
				continue
			}
			if clean {
				bot = *entities.CleanupBot(fakeRank, &bot)
			}
			actual = append(actual, bot)
		}
		return nil, actual
	}
}

func GetAllUsers(clean bool) (error, []entities.User) {
	redisUsers, err := Database.Redis.HVals("users").Result()
	if err != nil && len(redisUsers) > 0 {
		actual := make([]entities.User, len(redisUsers))
		for _, str := range redisUsers {
			user := entities.User{}
			err = Json.UnmarshalFromString(str, &user)
			if err != nil {
				continue
			}
			if clean {
				user = *entities.CleanupUser(fakeRank, &user)
			}
			actual = append(actual, user)
		}
		return nil, actual
	} else {
		cursor, err := Database.Mongo.Collection("users").Find(context.TODO(), bson.M{})
		if err != nil {
			return err, nil
		}
		var actual []entities.User
		defer cursor.Close(context.TODO())
		for cursor.Next(context.TODO()) {
			user := entities.User{}
			err = cursor.Decode(&user)
			if err != nil {
				continue
			}
			if clean {
				user = *entities.CleanupUser(fakeRank, &user)
			}
			actual = append(actual, user)
		}
		return nil, actual
	}
}

func GetAllServers(clean bool) (error, []entities.Server) {
	redisServers, err := Database.Redis.HVals("servers").Result()
	if err != nil && len(redisServers) > 0 {
		actual := make([]entities.Server, len(redisServers))
		for _, str := range redisServers {
			server := entities.Server{}
			err = Json.UnmarshalFromString(str, &server)
			if err != nil {
				continue
			}
			if clean {
				server = *entities.CleanupServer(fakeRank, &server)
			}
			actual = append(actual, server)
		}
		return nil, actual
	} else {
		cursor, err := Database.Mongo.Collection("bots").Find(context.TODO(), bson.M{})
		if err != nil {
			return err, nil
		}
		var actual []entities.Server
		defer cursor.Close(context.TODO())
		for cursor.Next(context.TODO()) {
			server := entities.Server{}
			err = cursor.Decode(&server)
			if err != nil {
				continue
			}
			if clean {
				server = *entities.CleanupServer(fakeRank, &server)
			}
			actual = append(actual, server)
		}
		return nil, actual
	}
}
