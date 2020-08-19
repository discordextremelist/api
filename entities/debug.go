package entities

import (
	"sync"
)

type ResponseTime struct {
	Path                 string `json:"path,omitempty"`
	TimeTakenToLookup    int64  `json:"time_taken_to_lookup"`
	TimeSpentDecoding    int64  `json:"time_spent_decoding"`
	TimeSpentWritingBody int64  `json:"time_spent_writing_body"`
}

type LookupTimes struct {
	Mongo map[string]map[string][]ResponseTime `json:"mongo"`
	Redis map[string]map[string][]ResponseTime `json:"redis"`
}

type DebugStatistics struct {
	MongoPing     int64          `json:"mongo_ping"`
	RedisPing     int64          `json:"redis_ping"`
	LookupTimes   LookupTimes    `json:"lookup_times"`
	ResponseTimes []ResponseTime `json:"response_times"`
	Hostname      string         `json:"hostname"`
}

var (
	MongoLookupTimes = make(map[string]map[string][]ResponseTime)
	RedisLookupTimes = make(map[string]map[string][]ResponseTime)
	ResponseTimes    []ResponseTime
	mutex            = &sync.Mutex{}
)

func AddMongoLookupTime(col, id string, lookup int64, unmarshal int64) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := MongoLookupTimes[col]; ok {
		if itemEntry, ok := MongoLookupTimes[col][id]; ok {
			itemEntry = append(itemEntry, ResponseTime{
				TimeTakenToLookup:    lookup,
				TimeSpentDecoding:    unmarshal,
				TimeSpentWritingBody: 0,
			})
			MongoLookupTimes[col][id] = itemEntry
		} else {
			MongoLookupTimes[col][id] = []ResponseTime{
				{
					TimeTakenToLookup:    lookup,
					TimeSpentDecoding:    unmarshal,
					TimeSpentWritingBody: 0,
				},
			}
		}
	} else {
		MongoLookupTimes[col] = map[string][]ResponseTime{
			id: {
				{
					TimeTakenToLookup:    lookup,
					TimeSpentDecoding:    unmarshal,
					TimeSpentWritingBody: 0,
				},
			},
		}
	}
}

func AddRedisLookupTime(col, id string, lookup int64, unmarshal int64) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := RedisLookupTimes[col]; ok {
		if itemEntry, ok := RedisLookupTimes[col][id]; ok {
			itemEntry = append(itemEntry, ResponseTime{
				TimeTakenToLookup:    lookup,
				TimeSpentDecoding:    unmarshal,
				TimeSpentWritingBody: -1,
			})
			RedisLookupTimes[col][id] = itemEntry
		} else {
			RedisLookupTimes[col][id] = []ResponseTime{
				{
					TimeTakenToLookup:    lookup,
					TimeSpentDecoding:    unmarshal,
					TimeSpentWritingBody: -1,
				},
			}
		}
	} else {
		RedisLookupTimes[col] = map[string][]ResponseTime{
			id: {
				{
					TimeTakenToLookup:    lookup,
					TimeSpentDecoding:    unmarshal,
					TimeSpentWritingBody: -1,
				},
			},
		}
	}
}
