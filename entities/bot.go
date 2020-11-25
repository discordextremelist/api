package entities

import (
	"context"
	"encoding/json"
	"github.com/discordextremelist/api/util"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
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
	Invite        string `json:"invite"`
	Support       string `json:"support"`
	Website       string `json:"website"`
	Donation      string `json:"donation"`
	Repo          string `json:"repo"`
	PrivacyPolicy string `json:"privacyPolicy"`
}

type BotSocial struct {
	Twitter string `json:"twitter"`
}

type BotTheme struct {
	UseCustomColour bool   `json:"useCustomColour"`
	Colour          string `json:"colour"`
	Banner          string `json:"banner"`
}

type WidgetBot struct {
	Channel string `json:"channel"`
	Options string `json:"options"`
	Server  string `json:"server"`
}

type Bot struct {
	MongoID     string     `json:"_id,omitempty"`
	ID          string     `bson:"_id" json:"id"`
	Name        string     `json:"name"`
	Prefix      string     `json:"prefix"`
	Library     string     `json:"library"`
	Tags        []string   `json:"tags"`
	VanityURL   string     `json:"vanityUrl"`
	ServerCount int        `json:"serverCount"`
	ShardCount  int        `json:"shardCount"`
	Token       string     `json:"token,omitempty"`
	Flags       int        `json:"flags"`
	ShortDesc   string     `json:"shortDesc"`
	LongDesc    string     `json:"longDesc"`
	ModNotes    string     `json:"modNotes,omitempty"`
	Editors     []string   `json:"editors"`
	Owner       Owner      `json:"owner"`
	Avatar      Avatar     `json:"avatar"`
	Votes       *BotVotes  `json:"votes,omitempty"`
	Links       BotLinks   `json:"links"`
	Social      BotSocial  `json:"social"`
	Theme       *BotTheme  `json:"theme,omitempty"`
	WidgetBot   *WidgetBot `json:"widgetbot,omitempty"`
	Status      BotStatus  `json:"status"`
}

func CleanupBot(rank UserRank, bot *Bot) *Bot {
	copied := *bot
	copied.ModNotes = ""
	copied.Token = ""
	copied.Votes = nil
	copied.Status.Premium = false
	copied.Theme = nil
	copied.WidgetBot = nil
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
	findStart := time.Now()
	var findEnd int64
	res := util.Database.Mongo.Collection("bots").FindOne(context.TODO(), bson.M{"_id": id})
	if res.Err() != nil {
		AddMongoLookupTime("bots", id, time.Since(findStart).Microseconds(), -1)
		return res.Err(), nil
	}
	findEnd = time.Since(findStart).Microseconds()
	bot := Bot{}
	decodeStart := time.Now()
	var decodeEnd int64
	if err := res.Decode(&bot); err != nil {
		AddMongoLookupTime("bots", id, findEnd, time.Since(decodeStart).Microseconds())
		return err, nil
	}
	decodeEnd = time.Since(decodeStart).Microseconds()
	AddMongoLookupTime("bots", id, findEnd, decodeEnd)
	return nil, &bot
}

func LookupBot(id string, clean bool) (error, *Bot) {
	findStart := time.Now()
	var findEnd int64
	redisBot, err := util.Database.Redis.HGet(context.TODO(), "bots", id).Result()
	if err == nil {
		findEnd = time.Since(findStart).Microseconds()
		if redisBot == "" {
			err, bot := mongoLookupBot(id)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					return err, nil
				}
				log.Errorf("Fallback for MongoDB failed for LookupBot(%s): %v", id, err.Error())
				return LookupError, nil
			} else {
				bot.MongoID = ""
				if clean {
					bot = CleanupBot(fakeRank, bot)
				}
				return nil, bot
			}
		}
		bot := &Bot{}
		decodeStart := time.Now()
		err = json.Unmarshal([]byte(redisBot), &bot)
		if err != nil {
			AddRedisLookupTime("bots", id, findEnd, time.Since(decodeStart).Microseconds())
			log.Errorf("Json parsing failed for LookupBot(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			if bot.ID == "" {
				bot.ID = bot.MongoID
				bot.MongoID = ""
			} else {
				bot.MongoID = ""
			}
			if clean {
				bot = CleanupBot(fakeRank, bot)
			}
			AddRedisLookupTime("bots", id, findEnd, time.Since(decodeStart).Microseconds())
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
			bot.MongoID = ""
			if clean {
				bot = CleanupBot(fakeRank, bot)
			}
			return nil, bot
		}
	}
}

func GetAllBots(clean bool) (error, []Bot) {
	redisBots, err := util.Database.Redis.HVals(context.TODO(), "bots").Result()
	var actual []Bot
	if err != nil {
		return err, nil
	}
	for _, str := range redisBots {
		bot := Bot{}
		err = json.Unmarshal([]byte(str), &bot)
		if err != nil {
			continue
		}
		if bot.ID == "" {
			bot.ID = bot.MongoID
			bot.MongoID = ""
		} else {
			bot.MongoID = ""
		}
		if clean {
			bot = *CleanupBot(fakeRank, &bot)
		}
		actual = append(actual, bot)
	}
	return nil, actual
}
