package entities

import (
	"context"
	"github.com/discordextremelist/api/util"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type BotStatus struct {
	Approved bool `json:"approved"`
	Premium  bool `json:"premium,omitempty"`
	SiteBot  bool `json:"siteBot"`
	Archived bool `json:"archived"`
}

type BotVotes struct {
	Positive []string `json:"positive"`
	Negative []string `json:"negative"`
}

type BotLinks struct {
	Invite   string `json:"invite"`
	Support  string `json:"support"`
	Website  string `json:"website"`
	Donation string `json:"donation"`
	Repo     string `json:"repo"`
}

type WidgetBot struct {
	Channel string `json:"channel"`
	Options string `json:"options"`
	Server  string `json:"server"`
}

type Bot struct {
	ID          string    `bson:"_id" json:"id"`
	Name        string    `json:"name"`
	Prefix      string    `json:"prefix"`
	Tags        []string  `json:"tags"`
	VanityURL   string    `json:"vanityUrl"`
	ServerCount int       `json:"serverCount"`
	ShardCount  int       `json:"shardCount"`
	Token       string    `json:"token,omitempty"`
	ShortDesc   string    `json:"shortDesc"`
	LongDesc    string    `json:"longDesc"`
	ModNotes    string    `json:"modNotes,omitempty"`
	Editors     []string  `json:"editors"`
	Owner       Owner     `json:"owner"`
	Avatar      Avatar    `json:"avatar"`
	Votes       *BotVotes `json:"votes,omitempty"`
	Links       BotLinks  `json:"links"`
	Status      BotStatus `json:"status"`
}

func CleanupBot(rank UserRank, bot *Bot) *Bot {
	copied := *bot
	copied.ModNotes = ""
	copied.Token = ""
	copied.Votes = nil
	copied.Status.Premium = false
	if rank.Mod {
		copied.ModNotes = bot.ModNotes
	}
	if rank.Admin || rank.Assistant {
		copied.ModNotes = bot.ModNotes
		copied.Token = bot.Token
		copied.Votes = bot.Votes
		copied.Status.Premium = bot.Status.Premium
	}
	return &copied
}

func mongoLookupBot(id string) (error, *Bot) {
	res := util.Database.Mongo.Collection("bots").FindOne(context.TODO(), bson.M{"_id": id})
	if res.Err() != nil {
		return res.Err(), nil
	}
	bot := Bot{}
	if err := res.Decode(&bot); err != nil {
		return err, nil
	}
	return nil, &bot
}

func LookupBot(id string, clean bool) (error, *Bot) {
	redisBot, err := util.Database.Redis.HGet(context.TODO(), "bots", id).Result()
	if err != nil {
		if redisBot == "" {
			err, bot := mongoLookupBot(id)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					return err, nil
				}
				log.Errorf("Fallback for MongoDB failed for LookupBot(%s): %v", id, err.Error())
				return LookupError, nil
			} else {
				if clean {
					bot = CleanupBot(fakeRank, bot)
				}
				return nil, bot
			}
		}
		bot := &Bot{}
		err = util.Json.UnmarshalFromString(redisBot, &bot)
		if err != nil {
			log.Errorf("Json parsing failed for LookupBot(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			if clean {
				bot = CleanupBot(fakeRank, bot)
			}
			return nil, bot
		}
	} else {
		err, bot := mongoLookupBot(id)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return err, nil
			}
			log.Errorf("Fallback for MongoDB failed for LookupBot(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			if clean {
				bot = CleanupBot(fakeRank, bot)
			}
			return nil, bot
		}
	}
}

func GetAllBots(clean bool) (error, []Bot) {
	redisBots, err := util.Database.Redis.HVals(context.TODO(), "bots").Result()
	if err != nil && len(redisBots) > 0 {
		actual := make([]Bot, len(redisBots))
		for _, str := range redisBots {
			bot := Bot{}
			err = util.Json.UnmarshalFromString(str, &bot)
			if err != nil {
				continue
			}
			if clean {
				bot = *CleanupBot(fakeRank, &bot)
			}
			actual = append(actual, bot)
		}
		return nil, actual
	} else {
		cursor, err := util.Database.Mongo.Collection("bots").Find(context.TODO(), bson.M{})
		if err != nil {
			return err, nil
		}
		var actual []Bot
		defer cursor.Close(context.TODO())
		for cursor.Next(context.TODO()) {
			bot := Bot{}
			err = cursor.Decode(&bot)
			if err != nil {
				continue
			}
			if clean {
				bot = *CleanupBot(fakeRank, &bot)
			}
			actual = append(actual, bot)
		}
		return nil, actual
	}
}
