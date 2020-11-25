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

type ServerLinks struct {
	Invite   string `json:"invite,omitempty"`
	Website  string `json:"website"`
	Donation string `json:"donation"`
}

type ServerStatus struct {
	ReviewRequired bool `json:"reviewRequired"`
}

type Server struct {
	MongoID        string       `json:"_id,omitempty"`
	ID             string       `bson:"_id" json:"id"`
	InviteCode     string       `json:"inviteCode,omitempty"`
	Name           string       `json:"name"`
	ShortDesc      string       `json:"shortDesc"`
	LongDesc       string       `json:"longDesc"`
	Tags           []string     `json:"tags"`
	PreviewChannel string       `json:"previewChannel"`
	Owner          Owner        `json:"owner"`
	Icon           Avatar       `json:"icon"`
	Links          ServerLinks  `json:"links"`
	Status         ServerStatus `json:"status"`
}

func CleanupServer(rank UserRank, server *Server) *Server {
	copied := *server
	copied.InviteCode = ""
	copied.Links.Invite = ""
	if rank.Admin || rank.Assistant {
		copied.InviteCode = server.InviteCode
		copied.Links.Invite = server.Links.Invite
	}
	return &copied
}

func mongoLookupServer(id string) (error, *Server) {
	findStart := time.Now()
	var findEnd int64
	res := util.Database.Mongo.Collection("servers").FindOne(context.TODO(), bson.M{"_id": id})
	if res.Err() != nil {
		AddMongoLookupTime("servers", id, time.Since(findStart).Microseconds(), -1)
		return res.Err(), nil
	}
	findEnd = time.Since(findStart).Microseconds()
	server := Server{}
	decodeStart := time.Now()
	var decodeEnd int64
	if err := res.Decode(&server); err != nil {
		AddMongoLookupTime("servers", id, findEnd, time.Since(decodeStart).Microseconds())
		return err, nil
	}
	decodeEnd = time.Since(decodeStart).Microseconds()
	AddMongoLookupTime("servers", id, findEnd, decodeEnd)
	return nil, &server
}

func LookupServer(id string, clean bool) (error, *Server) {
	findStart := time.Now()
	var findEnd int64
	redisServer, err := util.Database.Redis.HGet(context.TODO(), "servers", id).Result()
	if err == nil {
		findEnd = time.Since(findStart).Microseconds()
		if redisServer == "" {
			err, server := mongoLookupServer(id)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					return err, nil
				}
				log.Errorf("Fallback for MongoDB failed for LookupServer(%s): %v", id, err.Error())
				return LookupError, nil
			} else {
				server.MongoID = ""
				if clean {
					server = CleanupServer(fakeRank, server)
				}
				return nil, server
			}
		}
		server := &Server{}
		decodeStart := time.Now()
		err = json.Unmarshal([]byte(redisServer), &server)
		if err != nil {
			AddRedisLookupTime("servers", id, findEnd, time.Since(decodeStart).Microseconds())
			log.Errorf("Json parsing failed for LookupServer(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			if server.ID == "" {
				server.ID = server.MongoID
				server.MongoID = ""
			} else {
				server.MongoID = ""
			}
			if clean {
				server = CleanupServer(fakeRank, server)
			}
			AddRedisLookupTime("servers", id, findEnd, time.Since(decodeStart).Microseconds())
			return nil, server
		}
	} else {
		err, server := mongoLookupServer(id)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return err, nil
			}
			log.Errorf("Fallback for MongoDB failed for LookupServer(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			server.MongoID = ""
			if clean {
				server = CleanupServer(fakeRank, server)
			}
			return nil, server
		}
	}
}

func GetAllServers(clean bool) (error, []Server) {
	redisServers, err := util.Database.Redis.HVals(context.TODO(), "servers").Result()
	if err != nil {
		return err, nil
	}
	var actual []Server
	for _, str := range redisServers {
		server := Server{}
		err = json.Unmarshal([]byte(str), &server)
		if server.ID == "" {
			server.ID = server.MongoID
			server.MongoID = ""
		} else {
			server.MongoID = ""
		}
		if err != nil {
			continue
		}
		if clean {
			server = *CleanupServer(fakeRank, &server)
		}
		actual = append(actual, server)
	}
	return nil, actual
}
