package entities

import (
	"context"
	"encoding/json"
	"github.com/discordextremelist/api/util"
	"github.com/getsentry/sentry-go"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type UserPreferences struct {
	CustomGlobalCSS         string `json:"customGlobalCss"`
	DefaultColour           string `json:"defaultColour"`
	DefaultForegroundColour string `json:"defaultForegroundColour"`
	EnableGames             bool   `json:"enableGames"`
	Experiments             bool   `json:"experiments"`
}

type UserProfile struct {
	Bio   string           `json:"bio"`
	CSS   string           `json:"css,omitempty"`
	Links UserProfileLinks `json:"links"`
}

type UserGame struct {
	Snakes struct {
		MaxScore int `json:"maxScore"`
	} `json:"snakes"`
}

type UserRank struct {
	Admin      bool `json:"admin"`
	Assistant  bool `json:"assistant"`
	Mod        bool `json:"mod"`
	Premium    bool `json:"premium,omitempty"`
	Tester     bool `json:"tester"`
	Translator bool `json:"translator"`
	Covid      bool `json:"covid"`
}

type Strike struct {
	Reason   string `json:"reason"`
	Date     int    `json:"date"`
	Executor string `json:"executor"`
}

type SubHandled struct {
	Total      int `json:"total"`
	Approved   int `json:"approved"`
	Declined   int `json:"declined"`
	Unapproved int `json:"unapproved"`
	Removed    int `json:"removed"`
}

type StaffTracking struct {
	Details struct {
		Away struct {
			Status  bool   `json:"status"`
			Message string `json:"message"`
		} `json:"away"`
		Standing        string `json:"standing"`
		Country         string `json:"country"`
		Timezone        string `json:"timezone"`
		ManagementNotes string `json:"managementNotes"`
	} `json:"details"`
	LastLogin    int `json:"lastLogin"`
	LastAccessed struct {
		Time int    `json:"time"`
		Page string `json:"page"`
	} `json:"lastAccessed"`
	Punishments struct {
		Strikes  []Strike `json:"strikes"`
		Warnings []Strike `json:"warnings"`
	} `json:"punishments"`
	HandledBots struct {
		AllTime  SubHandled `json:"allTime"`
		PrevWeek SubHandled `json:"prevWeek"`
		ThisWeek SubHandled `json:"thisWeek"`
	} `json:"handledBots"`
	HandledServers struct {
		AllTime  SubHandled `json:"allTime"`
		PrevWeek SubHandled `json:"prevWeek"`
		ThisWeek SubHandled `json:"thisWeek"`
	} `json:"handledServers"`
	HandledTemplates struct {
		AllTime  SubHandled `json:"allTime"`
		PrevWeek SubHandled `json:"prevWeek"`
		ThisWeek SubHandled `json:"thisWeek"`
	} `json:"handledTemplates"`
}

type UserProfileLinks struct {
	Website   string `json:"website"`
	Github    string `json:"github"`
	Gitlab    string `json:"gitlab"`
	Twitter   string `json:"twitter"`
	Instagram string `json:"instagram"`
	Snapchat  string `json:"snapchat"`
}

type User struct {
	MongoID       string           `json:"_id,omitempty"`
	ID            string           `bson:"_id" json:"id"`
	Token         string           `json:"token,omitempty"`
	Name          string           `json:"name"`
	Discrim       string           `json:"discrim"`
	FullUsername  string           `json:"fullUsername"`
	Locale        string           `json:"locale,omitempty"`
	Flags         int              `json:"flags"`
	Avatar        Avatar           `json:"avatar"`
	Preferences   *UserPreferences `json:"preferences,omitempty"`
	Profile       UserProfile      `json:"profile"`
	Game          UserGame         `json:"game"`
	Rank          UserRank         `json:"rank"`
	StaffTracking *StaffTracking   `json:"staffTracking,omitempty"`
}

func CleanupUser(rank UserRank, user *User) *User {
	copied := *user
	copied.Locale = ""
	copied.Token = ""
	copied.Preferences = nil
	copied.Profile.CSS = ""
	copied.StaffTracking = nil
	copied.Rank.Premium = false
	if rank.Assistant || rank.Admin {
		copied.Locale = user.Locale
		copied.Preferences = user.Preferences
		copied.Profile.CSS = user.Profile.CSS
		copied.StaffTracking = user.StaffTracking
		copied.Rank.Premium = user.Rank.Premium
	}
	return &copied
}

func mongoLookupUser(id string) (error, *User) {
	findStart := time.Now()
	var findEnd int64
	res := util.Database.Mongo.Collection("users").FindOne(context.TODO(), bson.M{"_id": id})
	if res.Err() != nil {
		AddMongoLookupTime("users", id, time.Since(findStart).Microseconds(), -1)
		return res.Err(), nil
	}
	findEnd = time.Since(findStart).Microseconds()
	user := User{}
	decodeStart := time.Now()
	var decodeEnd int64
	if err := res.Decode(&user); err != nil {
		sentry.CaptureException(err)
		AddMongoLookupTime("users", id, findEnd, time.Since(decodeStart).Microseconds())
		return err, nil
	}
	decodeEnd = time.Since(decodeStart).Microseconds()
	AddMongoLookupTime("users", id, findEnd, decodeEnd)
	return nil, &user
}

func LookupUser(id string, clean bool) (error, *User) {
	findStart := time.Now()
	var findEnd int64
	redisUser, err := util.Database.Redis.HGet(context.TODO(), "users", id).Result()
	if err == nil {
		findEnd = time.Since(findStart).Microseconds()
		if redisUser == "" {
			err, user := mongoLookupUser(id)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					return err, nil
				}
				sentry.CaptureException(err)
				log.Errorf("Fallback for MongoDB failed for LookupUser(%s): %v", id, err.Error())
				return LookupError, nil
			} else {
				user.MongoID = ""
				if clean {
					user = CleanupUser(fakeRank, user)
				}
				return nil, user
			}
		}
		user := &User{}
		decodeStart := time.Now()
		err = json.Unmarshal([]byte(redisUser), &user)
		if err != nil {
			sentry.CaptureException(err)
			AddRedisLookupTime("users", id, findEnd, time.Since(decodeStart).Microseconds())
			log.Errorf("Json parsing failed for LookupUser(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			if user.ID == "" {
				user.ID = user.MongoID
				user.MongoID = ""
			} else {
				user.MongoID = ""
			}
			if clean {
				user = CleanupUser(fakeRank, user)
			}
			AddRedisLookupTime("users", id, findEnd, time.Since(decodeStart).Microseconds())
			return nil, user
		}
	} else {
		err, user := mongoLookupUser(id)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return err, nil
			}
			sentry.CaptureException(err)
			log.Errorf("Fallback for MongoDB failed for LookupUser(%s): %v", id, err.Error())
			return LookupError, nil
		} else {
			user.MongoID = ""
			if clean {
				user = CleanupUser(fakeRank, user)
			}
			return nil, user
		}
	}
}

func GetAllUsers(clean bool) (error, []User) {
	redisUsers := util.Scan[User]("users")
	var actual []User
	for _, user := range redisUsers {
		if user.ID == "" {
			user.ID = user.MongoID
			user.MongoID = ""
		} else {
			user.MongoID = ""
		}
		if clean {
			user = *CleanupUser(fakeRank, &user)
		}
		actual = append(actual, user)
	}
	return nil, actual
}
