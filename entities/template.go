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

type Role struct {
	Name        string `json:"name"`
	Color       int    `json:"color"`
	Hoist       bool   `json:"hoist"`
	Position    int    `json:"position"`
	Permissions int    `json:"permissions"`
	Managed     bool   `json:"managed"`
	Mentionable bool   `json:"mentionable"`
}

type PermissionsOverwrite struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // role/member
	Allow int    `json:"allow"`
	Deny  int    `json:"deny"`
}

type ServerTemplateLinks struct {
	LinkToServerPage	bool	`json:"linkToServerPage,omitempty"`
	Template 			string 	`json:"template"`
}

type GuildChannel struct {
	Type                  int                    `json:"type"`
	Position              int                    `json:"position,omitempty"`
	Name                  string                 `json:"name"`
	NSFW                  *bool                  `json:"nsfw,omitempty"`
	Topic                 *string                `json:"topic,omitempty"`
	PermissionsOverwrites []PermissionsOverwrite `json:"permissions_overwrites"`
	RateLimitPerUser      int                    `json:"rate_limit_per_user"`
	LastMessageID         string                 `json:"last_message_id"`
	Bitrate               *int                   `json:"bitrate,omitempty"`
	UserLimit             *int                   `json:"user_limit,omitempty"`
	LastPinTimestamp      *int                   `json:"last_pin_timestamp"`
}

type ServerTemplate struct {
	MongoID                     string              `json:"_id,omitempty"`
	ID                          string              `bson:"_id" json:"id"`
	Name                        string              `json:"name"`
	Region                      string              `json:"region"`
	Locale                      string              `json:"locale"`
	AfkTimeout                  int                 `json:"afkTimeout"`
	VerificationLevel           int                 `json:"verificationLevel"`
	DefaultMessageNotifications int                 `json:"defaultMessageNotifications"`
	ExplicitContent             int                 `json:"explicitContent"`
	Roles                       []Role              `json:"roles"`
	Channels                    []GuildChannel      `json:"channels"`
	UsageCount                  int                 `json:"usageCount"`
	ShortDesc                   string              `json:"shortDesc"`
	LongDesc                    string              `json:"longDesc"`
	Tags                        []string            `json:"tags"`
	FromGuild                   string              `json:"fromGuild"`
	Owner                       Owner               `json:"owner"`
	Icon                        Avatar              `json:"icon"`
	Links                       ServerTemplateLinks `json:"links"`
}

func mongoLookupTemplate(id string) (error, *ServerTemplate) {
	findStart := time.Now()
	var findEnd int64
	res := util.Database.Mongo.Collection("templates").FindOne(context.TODO(), bson.M{"_id": id})
	if res.Err() != nil {
		AddMongoLookupTime("templates", id, time.Since(findStart).Microseconds(), -1)
		return res.Err(), nil
	}
	findEnd = time.Since(findStart).Microseconds()
	template := ServerTemplate{}
	decodeStart := time.Now()
	var decodeEnd int64
	if err := res.Decode(&template); err != nil {
		AddMongoLookupTime("templates", id, findEnd, time.Since(decodeStart).Microseconds())
		return err, nil
	}
	decodeEnd = time.Since(decodeStart).Microseconds()
	AddMongoLookupTime("templates", id, findEnd, decodeEnd)
	return nil, &template
}

func LookupTemplate(id string) (error, *ServerTemplate) {
	findStart := time.Now()
	var findEnd int64
	redisTemplate, err := util.Database.Redis.HGet(context.TODO(), "templates", id).Result()
	if err == nil {
		findEnd = time.Since(findStart).Microseconds()
		if redisTemplate == "" {
			err, template := mongoLookupTemplate(id)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					return err, nil
				}
				log.Errorf("Fallback for MongoDB failed for LookupTemplate(%s): %v", id, err.Error())
				return LookupError, nil
			} else {
				template.MongoID = ""
				return nil, template
			}
		}
		template := &ServerTemplate{}
		decodeStart := time.Now()
		err = json.Unmarshal([]byte(redisTemplate), &template)
		if err != nil {
			AddRedisLookupTime("templates", id, findEnd, time.Since(decodeStart).Microseconds())
			log.Errorf("Json parsing failed for LookupTemplate(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			if template.ID == "" {
				template.ID = template.MongoID
				template.MongoID = ""
			} else {
				template.MongoID = ""
			}
			AddRedisLookupTime("templates", id, findEnd, time.Since(decodeStart).Microseconds())
			return nil, template
		}
	} else {
		err, template := mongoLookupTemplate(id)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return err, nil
			}
			log.Errorf("Fallback for MongoDB failed for LookupTemplate(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			template.MongoID = ""
			return nil, template
		}
	}
}

func GetAllTemplates() (error, []ServerTemplate) {
	redisTemplates, err := util.Database.Redis.HVals(context.TODO(), "templates").Result()
	if err == nil && len(redisTemplates) > 0 {
		var actual []ServerTemplate
		for _, str := range redisTemplates {
			template := ServerTemplate{}
			err = json.Unmarshal([]byte(str), &template)
			if template.ID == "" {
				template.ID = template.MongoID
				template.MongoID = ""
			} else {
				template.MongoID = ""
			}
			if err != nil {
				continue
			}
			actual = append(actual, template)
		}
		return nil, actual
	} else {
		cursor, err := util.Database.Mongo.Collection("templates").Find(context.TODO(), bson.M{})
		if err != nil {
			return err, nil
		}
		var actual []ServerTemplate
		defer cursor.Close(context.TODO())
		for cursor.Next(context.TODO()) {
			template := ServerTemplate{}
			err = cursor.Decode(&template)
			template.MongoID = ""
			if err != nil {
				continue
			}
			actual = append(actual, template)
		}
		return nil, actual
	}
}
