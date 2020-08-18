package entities

import (
	"context"
	"github.com/discordextremelist/api/util"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ServerLinks struct {
	Invite   string `json:"invite,omitempty"`
	Website  string `json:"website"`
	Donation string `json:"donation"`
}

type Server struct {
	ID         string      `bson:"_id" json:"id"`
	InviteCode string      `json:"inviteCode,omitempty"`
	Name       string      `json:"name"`
	ShortDesc  string      `json:"shortDesc"`
	LongDesc   string      `json:"longDesc"`
	Tags       []string    `json:"tags"`
	Owner      Owner       `json:"owner"`
	Icon       Avatar      `json:"icon"`
	Links      ServerLinks `json:"links"`
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
	res := util.Database.Mongo.Collection("servers").FindOne(context.TODO(), bson.M{"_id": id})
	if res.Err() != nil {
		return res.Err(), nil
	}
	server := Server{}
	if err := res.Decode(&server); err != nil {
		return err, nil
	}
	return nil, &server
}

func LookupServer(id string, clean bool) (error, *Server) {
	redisServer, err := util.Database.Redis.HGet(context.TODO(), "servers", id).Result()
	if err == nil {
		if redisServer == "" {
			err, server := mongoLookupServer(id)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					return err, nil
				}
				log.Errorf("Fallback for MongoDB failed for LookupServer(%s): %v", id, err.Error())
				return LookupError, nil
			} else {
				if clean {
					server = CleanupServer(fakeRank, server)
				}
				return nil, server
			}
		}
		server := &Server{}
		err = util.Json.UnmarshalFromString(redisServer, &server)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return err, nil
			}
			log.Errorf("Json parsing failed for LookupServer(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			if clean {
				server = CleanupServer(fakeRank, server)
			}
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
			if clean {
				server = CleanupServer(fakeRank, server)
			}
			return nil, server
		}
	}
}

func GetAllServers(clean bool) (error, []Server) {
	redisServers, err := util.Database.Redis.HVals(context.TODO(), "servers").Result()
	if err == nil && len(redisServers) > 0 {
		var actual []Server
		for _, str := range redisServers {
			server := Server{}
			err = util.Json.UnmarshalFromString(str, &server)
			if err != nil {
				continue
			}
			if clean {
				server = *CleanupServer(fakeRank, &server)
			}
			actual = append(actual, server)
		}
		return nil, actual
	} else {
		cursor, err := util.Database.Mongo.Collection("servers").Find(context.TODO(), bson.M{})
		if err != nil {
			return err, nil
		}
		var actual []Server
		defer cursor.Close(context.TODO())
		for cursor.Next(context.TODO()) {
			server := Server{}
			err = cursor.Decode(&server)
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
}
