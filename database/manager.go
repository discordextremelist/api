package database

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"strconv"
	"strings"
	"time"
)

type Manager struct {
	Redis *redis.Client
	Mongo *mongo.Database
}

func NewManager() Manager {
	return Manager{
		Redis: nil,
		Mongo: nil,
	}
}

func (manager *Manager) IsRedisOpen() bool {
	if err := manager.Redis.Ping(context.Background()).Err(); err != nil {
		return false
	}
	return true
}

func (manager *Manager) IsMongoOpen() bool {
	if err := manager.Mongo.Client().Ping(context.TODO(), readpref.Primary()); err != nil {
		return false
	}
	return true
}

func (manager *Manager) retryRedisConnect() {
	backoff := 0 * time.Second
	attempt := 0
	for {
		select {
		case <-time.After(backoff):
			{
				attempt++
				backoff = backoff + (1 * time.Second)
				if backoff.Seconds() >= 30 {
					backoff = 1 * time.Second
				}
				if err := manager.Redis.Ping(context.Background()).Err(); err != nil {
					log.WithField("type", "Redis").Warnf("Retry attempt %d failed!", attempt)
				} else {
					log.WithField("type", "Redis").Infof("Connected on attempt %d!", attempt)
					return
				}
			}
		}
	}
}

func (manager *Manager) OpenRedisConnection() {
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.WithField("type", "Redis").Fatal("Failed to convert type string to type int")
	}
	pass := os.Getenv("REDIS_PASSWORD")
	if sentinels := os.Getenv("REDIS_SENTINELS"); len(sentinels) > 0 {
		var splitSentinels = strings.Split(sentinels, ";")
		manager.Redis = redis.NewFailoverClient(&redis.FailoverOptions{
			SentinelAddrs: splitSentinels,
			MasterName:    os.Getenv("REDIS_MASTER"),
			Password:      pass,
			DB:            db,
			DialTimeout:   10 * time.Second,
			ReadTimeout:   15 * time.Second,
			WriteTimeout:  15 * time.Second,
		})
	} else {
		ip := os.Getenv("REDIS_IP")
		port := os.Getenv("REDIS_PORT")
		manager.Redis = redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:%s", ip, port),
			Password:     pass,
			DB:           db,
			DialTimeout:  10 * time.Second,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		})
	}
	manager.retryRedisConnect()
}

func (manager *Manager) OpenMongoConnection() {
	url := os.Getenv("MONGO_URL")
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(url))
	if err != nil {
		log.WithField("type", "MongoDB").Fatalf("Failed to connect to mongodb instance: %s", err.Error())
	}
	backoff := 0 * time.Second
	attempt := 0
	for {
		if attempt > 0 {
			backoff = backoff + (1 * time.Second)
		}
		if backoff.Seconds() > 30 {
			backoff = 1 * time.Second
		}
		attempt++
		if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
			log.WithField("type", "MongoDB").Warnf("Retry attempt %d failed!", attempt)
		} else {
			manager.Mongo = client.Database(os.Getenv("MONGO_DB"))
			log.WithField("type", "MongoDB").Infof("Connected on attempt %d!", attempt)
			return
		}
	}
}
